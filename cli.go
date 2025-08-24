package main

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/MatusOllah/slogcolor"
	"github.com/carapace-sh/carapace"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	PrintInit       = false
	PrintVersion    = false
	ToggleState     = false
	VerboseLog      = false
	Quiet           = false
	LyricOnly       = false
	Compact         = false
	Detailed        = false
	BreakTooltip    = 100000
	MaxTextLength   = 150
	TooltipLines    = 8
	TooltipColor    = "#cccccc"
	FilterProfanity = false
	LogFilePath     = ""

	FilterProfanityType = ""
)

func init() {
	var help bool
	Command.Flags().BoolVarP(&Compact, "compact", "c", Compact, "Output only text content on each line")
	Command.Flags().BoolVarP(&Detailed, "detailed", "d", Compact, "Put detailed player information in output (Experimental)")
	Command.Flags().BoolVarP(&LyricOnly, "lyric-only", "l", LyricOnly, "Display only lyrics in text output")
	Command.Flags().BoolVarP(&PrintInit, "init", "i", PrintInit, "Display JSON snippet for waybar/config.jsonc")
	Command.Flags().BoolVarP(&PrintVersion, "version", "V", PrintVersion, "Display waybar-lyric version information")
	Command.Flags().BoolVarP(&Quiet, "quiet", "q", Quiet, "Suppress all log output")
	Command.Flags().BoolVarP(&ToggleState, "toggle", "t", ToggleState, "Toggle player state between pause and resume")
	Command.Flags().BoolVarP(&VerboseLog, "verbose", "v", VerboseLog, "Enable verbose logging")
	Command.Flags().BoolVarP(&help, "help", "h", false, "Display help for waybar-lyric")
	Command.Flags().IntVarP(&BreakTooltip, "break-tooltip", "b", BreakTooltip, "Break long lines in tooltip")
	Command.Flags().IntVarP(&MaxTextLength, "max-length", "m", MaxTextLength, "Set maximum character length for lyrics text")
	Command.Flags().IntVarP(&TooltipLines, "tooltip-lines", "L", TooltipLines, "Set maximum number of lines in waybar tooltip")
	Command.Flags().StringVarP(&FilterProfanityType, "filter-profanity", "f", FilterProfanityType, "Filter profanity from lyrics (values: full, partial)")
	Command.Flags().StringVarP(&LogFilePath, "log-file", "o", LogFilePath, "Specify file path for saving logs")
	Command.Flags().StringVarP(&TooltipColor, "tooltip-color", "C", TooltipColor, "Set color for inactive lyrics lines")

	Command.Flags().MarkDeprecated("toggle", "use 'waybar-lyric play-pause'.")

	Command.AddCommand(playPauseCmd)

	comp := carapace.Gen(Command)
	comp.Standalone()
	comp.FlagCompletion(carapace.ActionMap{
		"log-file": carapace.ActionFiles(),
	})
}

var logFile *os.File

var Command = &cobra.Command{
	Use:          "waybar-lyric",
	Short:        "A waybar module for song lyrics",
	SilenceUsage: true,
	Run:          Execute,
	PostRunE: func(_ *cobra.Command, _ []string) error {
		if logFile != nil {
			return logFile.Close()
		}
		return nil
	},
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		switch FilterProfanityType {
		case "":
			FilterProfanity = false
		case "full", "partial":
			FilterProfanity = true
		default:
			return errors.New("Profanity filter must one of 'full' or 'partial'")
		}

		if TooltipLines < 4 {
			return errors.New("Tooltip lines limit must be at least 4")
		}

		if Quiet {
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

		if VerboseLog {
			opts.Level = slog.LevelDebug
		}

		if LogFilePath == "" {
			slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
			return nil
		}

		os.MkdirAll(filepath.Dir(LogFilePath), 0755)

		file, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
			slog.Error("Failed to open log-file", "error", err)
			return err
		}

		opts.NoColor = true
		slog.SetDefault(slog.New(slogcolor.NewHandler(file, opts)))
		logFile = file

		if ToggleState {
			cmd.RemoveCommand(playPauseCmd)
			return playPauseCmd.Execute()
		}

		return nil
	},
}
