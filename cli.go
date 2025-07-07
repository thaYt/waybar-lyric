package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
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
	Simplify        = false
	LogFilePath     = ""

	FilterProfanityType = ""
)

func init() {
	pflag.BoolVarP(&PrintInit, "init", "i", PrintInit, "Display JSON snippet for waybar/config.jsonc")
	pflag.BoolVarP(&PrintVersion, "version", "V", PrintVersion, "Display waybar-lyric version information")
	pflag.BoolVarP(&ToggleState, "toggle", "t", ToggleState, "Toggle player state between pause and resume")
	pflag.BoolVarP(&LyricOnly, "lyric-only", "l", LyricOnly, "Display only lyrics in text output")
	pflag.BoolVarP(&Compact, "compact", "c", Compact, "Output only text content on each line")
	pflag.BoolVarP(&Detailed, "detailed", "d", Compact, "Put detailed player information in output (Experimental)")
	pflag.BoolVarP(&Quiet, "quiet", "q", Quiet, "Suppress all log output")
	pflag.IntVarP(&BreakTooltip, "break-tooltip", "b", BreakTooltip, "Break long lines in tooltip")
	pflag.IntVarP(&MaxTextLength, "max-length", "m", MaxTextLength, "Set maximum character length for lyrics text")
	pflag.IntVarP(&TooltipLines, "tooltip-lines", "L", TooltipLines, "Set maximum number of lines in waybar tooltip")
	pflag.StringVarP(&TooltipColor, "tooltip-color", "C", TooltipColor, "Set color for inactive lyrics lines")
	pflag.StringVarP(&FilterProfanityType, "filter-profanity", "f", FilterProfanityType, "Filter profanity from lyrics (values: full, partial)")
	pflag.BoolVarP(&Simplify, "simplify", "s", Simplify, "lowercase + remove some other substitutions")
	pflag.BoolVarP(&VerboseLog, "verbose", "v", VerboseLog, "Enable verbose logging")
	pflag.StringVarP(&LogFilePath, "log-file", "o", LogFilePath, "Specify file path for saving logs")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprint(os.Stderr, "Get lyrics on waybar.\n\n")
		fmt.Println("Options:")
		fmt.Println(pflag.CommandLine.FlagUsages())
	}

	pflag.Parse()

	switch FilterProfanityType {
	case "":
		FilterProfanity = false
	case "full", "partial":
		FilterProfanity = true
	default:
		fmt.Fprintln(os.Stderr, "Profanity filter must one of 'full' or 'partial'")
		os.Exit(1)
	}

	if TooltipLines < 4 {
		fmt.Fprintln(os.Stderr, "Tooltip lines limit must be at least 4")
		return
	}

	if Quiet {
		slog.SetDefault(slog.New(&noopHandler{}))
		return
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
		return
	}

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
}
