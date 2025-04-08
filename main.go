package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/Pauloo27/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/pflag"
)

const (
	MaxLength  = 150
	SleepTime  = 500 * time.Millisecond
	PlayerName = "org.mpris.MediaPlayer2.spotify"
)

const Version = "waybar-lyric v0.6.2 (https://github.com/Nadim147c/waybar-lyric)"

func truncate(input string, limit int) string {
	r := []rune(input)

	if len(r) <= limit {
		return input
	}

	if limit > 3 {
		return string(r[:limit-3]) + "..."
	}

	return string(r[:limit])
}

func main() {
	init := pflag.Bool("init", false, "Show json snippet for waybar/config.jsonc")
	toggleState := pflag.Bool("toggle", false, "Toggle player state (pause/resume)")
	maxLineLength := pflag.Int("max-length", MaxLength, "Maximum lenght of lyrics text")
	version := pflag.Bool("version", false, "Print the version of waybar-lyric")
	logLevelF := pflag.BoolP("verbose", "v", false, "Use verbose loggin")
	logFile := pflag.String("log-file", "", "File to where logs should be save")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprint(os.Stderr, "Get spotify lyrics on waybar.\n\n")
		fmt.Println("Options:")
		fmt.Println(pflag.CommandLine.FlagUsages())
	}

	pflag.Parse()

	if *version {
		fmt.Fprint(os.Stderr, Version)
		return
	}

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
		fmt.Printf(`Put the following object in your waybar config:

"custom/lyrics": {
	"return-type": "json",
	"format": "{icon} {0}",
	"format-icons": {
		"playing": "",
		"paused": "",
		"lyric": "",
	},
	"exec-if": "which waybar-lyric",
	"exec": "waybar-lyric --max-length %d",
	"on-click": "waybar-lyric --toggle",
},
`, *maxLineLength)
		return
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		slog.Error("Failed to create dbus connection", "error", err)
		return
	}

	player := mpris.New(conn, PlayerName)

	if *toggleState {
		slog.Info("Toggling player state")
		if err := player.PlayPause(); err != nil {
			slog.Error("Failed to toggle player state", "error", err)
		}
		return
	}

	var lastInfo *PlayerInfo = nil
	var lastLine *LyricLine = nil
	playerOpened := true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	psChan := make(chan *dbus.Signal, 0)
	player.OnSignal(psChan)

	lyricTicker := time.NewTicker(SleepTime)
	defer lyricTicker.Stop()

	// Main loop
	fixedTicker := time.NewTicker(SleepTime)
	defer fixedTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Clean exit on cancel
		case <-psChan:
			slog.Debug("Received player update signal")
		case <-lyricTicker.C:
		case <-fixedTicker.C:
		}

		if _, err := player.GetPosition(); err != nil {
			if playerOpened {
				slog.Error("Player not found!", "error", err)
				fmt.Println("{}")
				playerOpened = false
			}
			continue
		} else {
			playerOpened = true
		}

		info, err := GetSpotifyInfo(player)
		if err != nil {
			slog.Error("Failed to parse dbus mpris metadata", "error", err)
			fmt.Println("{}")
			continue
		}

		playerUpdated := lastInfo == nil || lastInfo.ID != info.ID || lastInfo.Status != info.Status

		if playerUpdated {
			slog.Info("Player media found", "title", info.Title, "artist", info.Artist, "status", info.Status)
			lastInfo = info
		}

		if info.Status == mpris.PlaybackStopped {
			slog.Info("Player is stopped")
			fmt.Println("{}")
			continue
		}

		if info.Status == mpris.PlaybackPaused {
			if playerUpdated {
				info.Waybar().Encode()
				lastLine = nil
			}
			continue
		}

		lyrics, err := FetchLyrics(info)
		if err != nil {
			slog.Error("Failed to get lyrics", "error", err)
			info.Waybar().Encode()
			continue
		}

		var idx int
		for i, line := range lyrics {
			if info.Position <= line.Timestamp {
				break
			}
			idx = i
		}

		lyric := lyrics[idx]
		if lyric.Text != "" {
			if lastLine != nil && lastLine.Timestamp == lyric.Timestamp {
				continue
			}
			lastLine = &lyric

			slog.Info("Lyrics", "line", lyric.Text)
			NewWaybar(lyrics, idx, info.Percentage(), *maxLineLength).Encode()

			if len(lyrics) > idx+1 {
				n := lyrics[idx+1]
				d := n.Timestamp - info.Position
				slog.Debug("Sleep", "duration", d.String(), "position", info.Position.String(), "next", n.Timestamp.String())
				lyricTicker.Reset(d)
			}
			continue
		}

		if playerUpdated {
			info.Waybar().Encode()
		}
	}
}
