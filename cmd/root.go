package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/MatusOllah/slogcolor"
	initcmd "github.com/Nadim147c/waybar-lyric/cmd/init"
	"github.com/Nadim147c/waybar-lyric/cmd/playpause"
	"github.com/Nadim147c/waybar-lyric/cmd/position"
	"github.com/Nadim147c/waybar-lyric/cmd/seek"
	"github.com/Nadim147c/waybar-lyric/cmd/volume"
	"github.com/Nadim147c/waybar-lyric/internal/config"
	"github.com/carapace-sh/carapace"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	Command.Flags().BoolVarP(&config.Compact, "compact", "c", config.Compact, "Output only text content on each line")
	Command.Flags().BoolVarP(&config.Detailed, "detailed", "d", config.Detailed, "Put detailed player information in output (Experimental)")
	Command.Flags().BoolVarP(&config.LyricOnly, "lyric-only", "l", config.LyricOnly, "Display only lyrics in text output")
	Command.Flags().BoolVarP(&config.PrintInit, "init", "i", config.PrintInit, "Display JSON snippet for waybar/config.jsonc")
	Command.Flags().BoolVarP(&config.PrintVersion, "version", "V", config.PrintVersion, "Display waybar-lyric version information")
	Command.Flags().BoolVarP(&config.ToggleState, "toggle", "t", config.ToggleState, "Toggle player state between pause and resume")
	Command.Flags().IntVarP(&config.BreakTooltip, "break-tooltip", "b", config.BreakTooltip, "Break long lines in tooltip")
	Command.Flags().IntVarP(&config.MaxTextLength, "max-length", "m", config.MaxTextLength, "Set maximum character length for lyrics text")
	Command.Flags().IntVarP(&config.TooltipLines, "tooltip-lines", "L", config.TooltipLines, "Set maximum number of lines in waybar tooltip")
	Command.Flags().StringVarP(&config.FilterProfanityType, "filter-profanity", "f", config.FilterProfanityType, "Filter profanity from lyrics (values: full, partial)")
	Command.Flags().StringVarP(&config.TooltipColor, "tooltip-color", "C", config.TooltipColor, "Set color for inactive lyrics lines")
	Command.Flags().BoolVarP(&config.Simplify, "simplify", "s", config.Simplify, "lowercase + remove some other substitutions")

	Command.Flags().MarkDeprecated("init", "use 'waybar-lyric init'.")
	Command.Flags().MarkDeprecated("toggle", "use 'waybar-lyric play-pause'.")

	Command.MarkFlagsMutuallyExclusive("toggle", "init")

	Command.PersistentFlags().BoolP("help", "h", false, "Display help for waybar-lyric")
	Command.PersistentFlags().BoolVarP(&config.Quiet, "quiet", "q", config.Quiet, "Suppress all log output")
	Command.PersistentFlags().BoolVarP(&config.VerboseLog, "verbose", "v", config.VerboseLog, "Enable verbose logging")
	Command.PersistentFlags().StringVarP(&config.LogFilePath, "log-file", "o", config.LogFilePath, "Specify file path for saving logs")

	Command.AddCommand(initcmd.Command)
	Command.AddCommand(playpause.Command)
	Command.AddCommand(position.Command)
	Command.AddCommand(seek.Command)
	Command.AddCommand(volume.Command)

	comp := carapace.Gen(Command)
	comp.Standalone()
	comp.FlagCompletion(carapace.ActionMap{
		"log-file": carapace.ActionFiles(),
	})
}

var logFile *os.File

// Command is root command for waybar
var Command = &cobra.Command{
	Use:          "waybar-lyric",
	Short:        "A waybar module for song lyrics",
	SilenceUsage: true,
	Run:          Execute,
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if logFile != nil {
			return logFile.Close()
		}
		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if config.ToggleState {
			defer func() {
				cmd.RemoveCommand(playpause.Command)
				playpause.Command.Execute()
				os.Exit(0)
			}()
		}

		if config.PrintInit {
			defer func() {
				cmd.RemoveCommand(initcmd.Command)
				initcmd.Command.Execute()
				os.Exit(0)
			}()
		}

		switch config.FilterProfanityType {
		case "":
			config.FilterProfanity = false
		case "full", "partial":
			config.FilterProfanity = true
		default:
			return errors.New("Profanity filter must one of 'full' or 'partial'")
		}

		if config.TooltipLines < 4 {
			return errors.New("Tooltip lines limit must be at least 4")
		}

		if config.Quiet {
			slog.SetDefault(slog.New(&noopHandler{}))
			return nil
		}

		opts := slogcolor.DefaultOptions
		opts.LevelTags = map[slog.Level]string{
			slog.LevelDebug: color.New(color.FgGreen).Sprint("DEBUG"),
			slog.LevelInfo:  color.New(color.FgCyan).Sprint("INFO "),
			slog.LevelWarn:  color.New(color.FgYellow).Sprint("WARN "),
			slog.LevelError: color.New(color.FgRed).Sprint("ERROR"),
		}

		if config.VerboseLog {
			opts.Level = slog.LevelDebug
		}

		if config.LogFilePath == "" {
			slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
			return nil
		}

		os.MkdirAll(filepath.Dir(config.LogFilePath), 0755)

		file, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
			slog.Error("Failed to open log-file", "error", err)
			return err
		}

		opts.NoColor = true
		slog.SetDefault(slog.New(slogcolor.NewHandler(file, opts)))
		logFile = file

		return nil
	},
}

// noopHandler is a empty handler that ignore all logs
type noopHandler struct{}

var _ slog.Handler = (*noopHandler)(nil)

func (h *noopHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (h *noopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return h }
func (h *noopHandler) WithGroup(_ string) slog.Handler               { return h }
