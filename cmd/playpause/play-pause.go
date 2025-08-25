package playpause

import (
	"log/slog"

	"github.com/Nadim147c/go-mpris"
	"github.com/Nadim147c/waybar-lyric/internal/player"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

func init() {
	// This flag for ignoring by the playPauseCmd when running for deprecreated
	// waybar-lyric --toggle.
	Command.Flags().BoolP("toggle", "t", false, "Toggle player state between pause and resume")
	Command.Flags().MarkHidden("toggle")
}

// Command is the play-pause command
var Command = &cobra.Command{
	Use:          "play-pause",
	Short:        "Toggle play-pause state",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		conn, err := dbus.SessionBus()
		if err != nil {
			slog.Error("Failed to create dbus connection", "error", err)
			return err
		}

		var mp *mpris.Player
		for mp == nil {
			p, _, err := player.Select(conn)
			if err == nil {
				slog.Debug("Failed to select player", "error", err)
				mp = p
			}
		}
		slog.Debug("Player selected", "player", mp)

		slog.Info("Toggling player state")
		if err := mp.PlayPause(); err != nil {
			slog.Error("Failed to toggle player state", "error", err)
			return err
		}
		return nil
	},
}
