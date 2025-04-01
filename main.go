package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Pauloo27/go-mpris"
	"github.com/fred1268/go-clap/clap"
	"github.com/godbus/dbus/v5"
)

type Cli struct {
	Init      bool `clap:"--init"`
	Toggle    bool `clap:"--toggle"`
	MaxLength int  `clap:"--max-length"`
	Help      bool `clap:"--help,-h"`
}

func main() {
	args := make([]string, 0)
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	var err error
	cli := Cli{MaxLength: 150}
	if _, err = clap.Parse(args, &cli); err != nil {
		Log(err)
		os.Exit(1)
	}

	if cli.Help {
		fmt.Print(`Get spotify lyrics in your waybar

waybar-lyric [FLAGS]

  -h, --help          Show the help message
  --init              Show json snippet for waybar/config.jsonc
  --toggle            Pause or Resume spotify playback
  --max-length <int>  Maximum lenght of lyrics text
`)

		os.Exit(0)
	}

	Execute(cli)
}

func Log(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
}

func truncate(input string, limit int) string {
	if len(input) <= limit {
		return input
	}

	if limit > 3 {
		return input[:limit-3] + "..."
	}

	return input[:limit]
}

func Execute(cli Cli) {
	if cli.Init {
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
		Log("Failed to open or create lock file:", err)
		os.Exit(1)
	}
	defer file.Close()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			Log("Another instance of the CLI is already running. Exiting.")
			os.Exit(0)
		}
		Log("Failed to acquire lock:", err)
		os.Exit(1)
	}

	defer os.Remove(lockFile)

	conn, err := dbus.SessionBus()
	if err != nil {
		Log(err)
		os.Exit(1)
	}

	names, err := mpris.List(conn)
	if err != nil {
		Log(err)
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
		Log("failed to find player")
		os.Exit(1)
	}

	player := mpris.New(conn, playerName)

	if cli.Toggle {
		player.PlayPause()
		UpdateWaybar()
		os.Exit(0)
	}

	info, err := GetSpotifyInfo(player)
	if err != nil {
		Log(err)
		return
	}

	if info.Status == "Stopped" {
		os.Exit(0)
	}

	lyrics, err := FetchLyrics(info)
	if err != nil {
		Log(err)
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

		line := truncate(currentLine, cli.MaxLength)
		NewWaybarLyrics(line, tooltip.String(), info.Percentage()).Encode()

		os.Exit(0)
	}

	info.Waybar().Encode()
}
