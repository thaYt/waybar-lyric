package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	UserAgent      = "waybar-lyric v0.4.0 (https://github.com/Nadim147c/waybar-lyric)"
	LrclibEndpoint = "https://lrclib.net/api/get"
)

func request(params url.Values, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, LrclibEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = params.Encode()
	req.Header = header

	slog.Info("Fetching lyrics from Lrclib", "url", req.URL.String())

	client := http.Client{}

	return client.Do(req)
}

func FetchLyrics(info *PlayerInfo) ([]LyricLine, error) {
	uri := filepath.Base(info.ID)

	notFoundTempDir := filepath.Join(os.TempDir(), "waybar-lyric")
	lyricsNotFoundFile := filepath.Join(notFoundTempDir, uri+"-not-found")

	if _, err := os.Stat(lyricsNotFoundFile); err == nil {
		return nil, fmt.Errorf("Lyrics not found (cached)")
	}

	userCacheDir, _ := os.UserCacheDir()
	cacheDir := filepath.Join(userCacheDir, "waybar-lyric")

	uri = strings.ReplaceAll(uri, "/", "-")
	cacheFile := filepath.Join(cacheDir, uri+".csv")

	if cahcedLyrics, err := LoadCache(cacheFile); err == nil {
		return cahcedLyrics, nil
	} else {
		slog.Warn("Can't find the lyrics in the cache", "error", err)
	}

	queryParams := url.Values{}
	queryParams.Set("track_name", info.Title)
	queryParams.Set("artist_name", info.Artist)
	if info.Album != "" {
		queryParams.Set("album_name", info.Album)
	}
	if info.Length != 0 {
		queryParams.Set("duration", fmt.Sprintf("%.2f", info.Length.Seconds()))
	}

	header := http.Header{}
	header.Set("User-Agent", UserAgent)

	resp, err := request(queryParams, header)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		os.WriteFile(lyricsNotFoundFile, []byte(resp.Request.URL.String()), 644)
		return nil, fmt.Errorf("Lyrics not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	var resJson LrcLibResponse
	err = json.NewDecoder(resp.Body).Decode(&resJson)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	lyrics, err := parseLyrics(resJson.SyncedLyrics)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lyrics: %w", err)
	}

	if len(lyrics) == 0 {
		return nil, fmt.Errorf("failed to find sync lyrics lines")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	err = SaveCache(lyrics, cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to cache lyrics to psudo csv: %w", err)
	}

	return lyrics, nil
}
