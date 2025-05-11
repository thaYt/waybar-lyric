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
	MaxTextLength   = 150
	TooltipLines    = 8
	TooltipColor    = "#cccccc"
	FilterProfanity = false
	LogFilePath     = ""

	FilterProfanityType = ""
)

func init() {
	pflag.BoolVar(&PrintInit, "init", PrintInit, "Show JSON snippet for waybar/config.jsonc")
	pflag.BoolVar(&PrintVersion, "version", PrintVersion, "Print the version of waybar-lyric")
	pflag.BoolVar(&ToggleState, "toggle", ToggleState, "Toggle player state (pause/resume)")
	pflag.BoolVar(&LyricOnly, "lyric-only", LyricOnly, "Only show lyrics on text (Experimental)")
	pflag.BoolVar(&Compact, "compact", Compact, "Prints only the text in each line")
	pflag.BoolVar(&Quiet, "quiet", Quiet, "Avoid printing logs")
	pflag.IntVar(&MaxTextLength, "max-length", MaxTextLength, "Maximum length of lyrics text")
	pflag.IntVar(&TooltipLines, "tooltip-lines", TooltipLines, "Maximum lines of waybar tooltip")
	pflag.StringVarP(&TooltipColor, "tooltip-color", "t", TooltipColor, "Color of inactive lyrics lines")
	pflag.StringVarP(&FilterProfanityType, "filter-profanity", "f", FilterProfanityType, "Filter profanity from lyrics (full or partial)")
	pflag.BoolVarP(&VerboseLog, "verbose", "v", VerboseLog, "Use verbose logging")
	pflag.StringVar(&LogFilePath, "log-file", LogFilePath, "File where logs should be saved")

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
