package main

import (
	"log/slog"

	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

func init() {
	// This flag for ignoring by the playPauseCmd when running for deprecreated
	// waybar-lyric --toggle.
	playPauseCmd.Flags().BoolP("toggle", "t", false, "Toggle player state between pause and resume")
	playPauseCmd.Flags().MarkHidden("toggle")
}

var playPauseCmd = &cobra.Command{
	Use:          "play-pause",
	Short:        "Toggle play-pause state",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		conn, err := dbus.SessionBus()
		if err != nil {
			slog.Error("Failed to create dbus connection", "error", err)
			return err
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

		slog.Info("Toggling player state")
		if err := player.PlayPause(); err != nil {
			slog.Error("Failed to toggle player state", "error", err)
			return err
		}
		return nil
	},
}
