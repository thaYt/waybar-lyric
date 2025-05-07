package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
)

const (
	SleepTime = 500 * time.Millisecond
	Version   = "waybar-lyric v0.10.0 (https://github.com/Nadim147c/waybar-lyric)"
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
	if PrintVersion {
		fmt.Fprint(os.Stderr, Version)
		return
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

	var player *mpris.Player
	for player == nil {
		p, _, err := SelectPlayer(conn)
		if err == nil {
			slog.Debug("Failed to select player", "error", err)
			player = p
		}
	}
	slog.Debug("Player selected", "player", player)

	if ToggleState {
		slog.Info("Toggling player state")
		if err := player.PlayPause(); err != nil {
			slog.Error("Failed to toggle player state", "error", err)
		}
		return
	}

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

	var lastWaybar *Waybar = nil

	for {
		select {
		case <-ctx.Done():
			return // Clean exit on cancel
		case <-psChan:
			slog.Debug("Received player update signal")
		case <-lyricTicker.C:
		case <-fixedTicker.C:
		}

		player, parser, err := SelectPlayer(conn)
		if err != nil {
			slog.Error("Player not found!", "error", err)

			waybar := &Waybar{}
			if !waybar.Is(lastWaybar) {
				waybar.Encode()
				lastWaybar = waybar
			}

			continue
		}

		info, err := parser(player)
		if err != nil {
			slog.Error("Failed to parse dbus mpris metadata", "error", err)
			waybar := &Waybar{}
			if !waybar.Is(lastWaybar) {
				waybar.Encode()
				lastWaybar = waybar
			}
			continue
		}

		slog.Debug("PlayerInfo",
			"id", info.ID,
			"title", info.Title,
			"artist", info.Artist,
			"album", info.Album,
			"position", info.Position.String(),
			"length", info.Length.String(),
		)

		if info.Status == mpris.PlaybackStopped {
			slog.Info("Player is stopped")
			waybar := &Waybar{}
			if !waybar.Is(lastWaybar) {
				waybar.Encode()
				lastWaybar = waybar
			}
			continue
		}

		lyrics, err := GetLyrics(info)
		if err != nil {
			slog.Error("Failed to get lyrics", "error", err)
			waybar := info.Waybar()
			if !waybar.Is(lastWaybar) {
				waybar.Encode()
				lastWaybar = waybar
			}

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
			var tooltip strings.Builder
			tooltip.WriteString("<b><big>󰝚 </big></b>\n")
			tooltip.WriteString(fmt.Sprintf("<span foreground=\"%s\">", TooltipColor))

			end := min(TooltipLines, len(lyrics))
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

			if info.Status == mpris.PlaybackPaused {
				waybar.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
				waybar.Alt = Paused
				waybar.Class = Class{Paused}
				if !waybar.Is(lastWaybar) {
					waybar.Encode()
					lastWaybar = waybar
				}
				continue
			}

			if !waybar.Is(lastWaybar) {
				waybar.Encode()
				lastWaybar = waybar
			}
		} else {
			lyric := lyrics[idx]

			waybar := NewWaybar(lyrics, idx)

			if info.Status == mpris.PlaybackPaused {
				waybar.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
				waybar.Alt = Paused
				waybar.Class = Class{Paused}
				if !waybar.Is(lastWaybar) {
					slog.Info("Lyrics", "line", lyric.Text)
					waybar.Encode()
					lastWaybar = waybar
				}
				continue
			}

			if lyric.Text == "" {
				waybar.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
				waybar.Alt = Music
			}

			if !waybar.Is(lastWaybar) {
				slog.Info("Lyrics", "line", lyric.Text)
				waybar.Encode()
				lastWaybar = waybar
			}

			if len(lyrics) > idx+1 {
				n := lyrics[idx+1]
				d := n.Timestamp - info.Position
				if d <= 0 {
					slog.Warn("Negative sleep time",
						"duration", d.String(),
						"position", info.Position.String(),
						"next", n.Timestamp.String(),
					)
					continue
				}
				slog.Debug("Sleep",
					"duration", d.String(),
					"position", info.Position.String(),
					"next", n.Timestamp.String(),
				)
				lyricTicker.Reset(d)
			}
		}

	}
}
