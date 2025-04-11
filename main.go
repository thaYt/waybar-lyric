package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/pflag"
)

const (
	SleepTime  = 500 * time.Millisecond
	PlayerName = "org.mpris.MediaPlayer2.spotify"
	Version    = "waybar-lyric v0.6.3 (https://github.com/Nadim147c/waybar-lyric)"
)

var (
	PrintInit     = false
	PrintVersion  = false
	ToggleState   = false
	VerboseLog    = false
	MaxTextLength = 150
	TootlipColor  = "#cccccc"
	LogFilePath   = ""
)

func truncate(input string) string {
	r := []rune(input)

	if len(r) <= MaxTextLength {
		return input
	}

	if MaxTextLength > 3 {
		return string(r[:MaxTextLength-3]) + "..."
	}

	return string(r[:MaxTextLength])
}

func main() {
	pflag.BoolVar(&PrintInit, "init", PrintInit, "Show JSON snippet for waybar/config.jsonc")
	pflag.BoolVar(&PrintVersion, "version", PrintVersion, "Print the version of waybar-lyric")
	pflag.BoolVar(&ToggleState, "toggle", ToggleState, "Toggle player state (pause/resume)")
	pflag.IntVar(&MaxTextLength, "max-length", MaxTextLength, "Maximum length of lyrics text")
	pflag.StringVarP(&TootlipColor, "tooltip-color", "t", TootlipColor, "Maximum length of lyrics text")
	pflag.BoolVarP(&VerboseLog, "verbose", "v", VerboseLog, "Use verbose logging")
	pflag.StringVar(&LogFilePath, "log-file", LogFilePath, "File where logs should be saved")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprint(os.Stderr, "Get spotify lyrics on waybar.\n\n")
		fmt.Println("Options:")
		fmt.Println(pflag.CommandLine.FlagUsages())
	}

	pflag.Parse()

	if PrintVersion {
		fmt.Fprint(os.Stderr, Version)
		return
	}

	opts := slogcolor.DefaultOptions
	if VerboseLog {
		opts.Level = slog.LevelDebug
	}

	if LogFilePath != "" {
		os.MkdirAll(filepath.Dir(LogFilePath), 0755)

		file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
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

	if PrintInit {
		PrintSnippet()
		return
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		slog.Error("Failed to create dbus connection", "error", err)
		return
	}

	player := mpris.New(conn, PlayerName)

	if ToggleState {
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

		idx := -1
		for i, line := range lyrics {
			if info.Position <= line.Timestamp {
				break
			}
			idx = i
		}

		if idx == -1 {
			if lastLine != nil && lastLine.Timestamp == -1 {
				continue
			}
			lastLine = &LyricLine{Timestamp: -1, Text: ""}

			var tooltip strings.Builder
			tooltip.WriteString("<b><big>󰝚 </big></b>\n")
			tooltip.WriteString(fmt.Sprintf("<span foreground=\"%s\">", TootlipColor))

			end := min(5, len(lyrics))
			tooltipLyrics := lyrics[:end]
			for _, ttl := range tooltipLyrics {
				if ttl.Text != "" {
					tooltip.WriteString(ttl.Text + "\n")
				} else {
					tooltip.WriteString("󰝚 \n")
				}
			}

			waybar := info.Waybar()
			waybar.Tooltip = strings.TrimSpace(tooltip.String()) + "</span>"
			waybar.Alt = Music
			waybar.Class = Class{Playing, Music}
			waybar.Encode()
		} else {
			lyric := lyrics[idx]
			if lastLine != nil && lastLine.Timestamp == lyric.Timestamp {
				continue
			}
			lastLine = &lyric

			slog.Info("Lyrics", "line", lyric.Text)

			waybar := NewWaybar(lyrics, idx, info.Percentage())
			if lyric.Text != "" {
				waybar.Encode()
			} else {
				waybar.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
				waybar.Alt = Music
				waybar.Encode()
			}

			if len(lyrics) > idx+1 {
				n := lyrics[idx+1]
				d := n.Timestamp - info.Position
				slog.Debug("Sleep", "duration", d.String(), "position", info.Position.String(), "next", n.Timestamp.String())
				lyricTicker.Reset(d)
			}
		}

	}
}
