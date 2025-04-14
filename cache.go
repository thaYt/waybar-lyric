package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var CacheDir string

func init() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("Failed to find cache directory", "error", err)
		return
	}

	CacheDir = filepath.Join(userCacheDir, "waybar-lyric")

	if err := os.MkdirAll(CacheDir, 0755); err != nil {
		slog.Error("Failed to create cache directory")
	}
}

func SaveCache(lines []LyricLine, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for line := range slices.Values(lines) {
		_, err := fmt.Fprintf(file, "%d,%s\n", line.Timestamp, line.Text)
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadCache(filePath string) ([]LyricLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lyrics []LyricLine
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		ts, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		timestamp := time.Duration(ts)
		text := strings.TrimSpace(parts[1])

		lyric := LyricLine{Timestamp: timestamp, Text: text}
		lyrics = append(lyrics, lyric)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lyrics) == 0 {
		return nil, fmt.Errorf("Number of line found is zero.")
	}

	return lyrics, nil
}
