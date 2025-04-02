package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/MatusOllah/slogcolor"
	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/pflag"
)

const DefaultMaxLength = 150

func truncate(input string, limit int) string {
	if len(input) <= limit {
		return input
	}

	if limit > 3 {
		return input[:limit-3] + "..."
	}

	return input[:limit]
}

func main() {
	init := pflag.Bool("init", false, "Show json snippet for waybar/config.jsonc")
	toggleState := pflag.Bool("toggle", false, "Toggle player state (pause/resume)")
	maxLineLength := pflag.Int("max-length", DefaultMaxLength, "Maximum lenght of lyrics text")
	logLevelF := pflag.BoolP("verbose", "v", false, "Use verbose loggin")
	logFile := pflag.String("log-file", "", "File to where logs should be save")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprint(os.Stderr, "Get spotify lyrics on waybar.\n\n")
		fmt.Println("Options:")
		fmt.Println(pflag.CommandLine.FlagUsages())
	}

	pflag.Parse()

	opts := slogcolor.DefaultOptions
	if *logLevelF {
		opts.Level = slog.LevelDebug
	}

	if *logFile != "" {
		os.MkdirAll(filepath.Dir(*logFile), 0755)

		file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
			slog.Error("Failed to open log-file", "error", err)
		} else {
			opts.NoColor = true
			slog.SetDefault(slog.New(slogcolor.NewHandler(file, opts)))
			defer file.Close() // Close the file when done
		}
	} else {
		slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	}

	if *init {
		fmt.Print(`Put the following object in your waybar config:

"custom/lyrics": {
	"interval": 1,
	"signal": 4,
	"return-type": "json",
	"format": "{icon} {0}",
	"format-icons": {
		"playing": "",
		"paused": "",
		"lyric": "",
	},
	"exec-if": "which waybar-lyric",
	"exec": "waybar-lyric --max-length 100",
	"on-click": "waybar-lyric --toggle",
},
`)
		os.Exit(0)
	}

	lockFile := filepath.Join(os.TempDir(), "waybar-lyric.lock")
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		slog.Error("Failed to open or create lock file", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			slog.Warn("Another instance of the CLI is already running. Exiting.")
			os.Exit(0)
		}
		slog.Error("Failed to acquire lock", "error", err)
		os.Exit(1)
	}

	defer os.Remove(lockFile)

	conn, err := dbus.SessionBus()
	if err != nil {
		slog.Error("Failed to create dbus connection", "error", err)
		os.Exit(1)
	}

	names, err := mpris.List(conn)
	if err != nil {
		slog.Error("Failed to find list of player", "error", err)
		os.Exit(1)
	}

	searchTerm := "spotify"
	var playerName string
	for _, name := range names {
		if strings.Contains(strings.ToLower(name), strings.ToLower(searchTerm)) {
			playerName = name
			break
		}
	}

	if playerName == "" {
		slog.Error("Can't find supported player", "error", err)
		os.Exit(1)
	}

	player := mpris.New(conn, playerName)

	if *toggleState {
		slog.Info("Toggling player state")
		if err := player.PlayPause(); err != nil {
			slog.Error("Failed to toggle player state", "error", err)
		}

		if err := UpdateWaybar(); err != nil {
			slog.Error("Failed to update waybar through signals", "error", err)
		}
		os.Exit(0)
	}

	info, err := GetSpotifyInfo(player)
	if err != nil {
		slog.Error("Failed to parse dbus mpris metadata", "error", err)
		return
	}

	slog.Info("Player media found", "title", info.Title, "artist", info.Artist)

	if info.Status == mpris.PlaybackStopped {
		slog.Info("Player is stopped")
		os.Exit(0)
	}

	if info.Status == mpris.PlaybackPaused {
		info.Waybar().Encode()
		os.Exit(0)
	}

	lyrics, err := FetchLyrics(info)
	if err != nil {
		slog.Error("Failed to get lyrics", "error", err)
		info.Waybar().Encode()
		os.Exit(0)
	}

	var idx int
	for i, line := range lyrics {
		if info.Position < line.Timestamp {
			break
		}
		idx = i
	}

	currentLine := lyrics[idx].Text

	if currentLine != "" {
		start := max(idx-2, 0)
		end := min(idx+5, len(lyrics))

		tooltipLyrics := lyrics[start:end]
		var tooltip strings.Builder
		for i, ttl := range tooltipLyrics {
			lineText := ttl.Text
			if start+i == idx {
				tooltip.WriteString("> ")
			}
			tooltip.WriteString(lineText + "\n")
		}

		line := truncate(currentLine, *maxLineLength)
		NewWaybarLyrics(line, tooltip.String(), info.Percentage()).Encode()

		os.Exit(0)
	}

	info.Waybar().Encode()
}
