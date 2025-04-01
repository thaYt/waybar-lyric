package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const LrclibEndpoint = "https://lrclib.net/api/get"

func FetchLyrics(info *PlayerInfo) ([]LyricLine, error) {
	queryParams := url.Values{}
	queryParams.Set("track_name", info.Title)
	queryParams.Set("artist_name", info.Artist)
	if info.Album != "" {
		queryParams.Set("album_name", info.Album)
	}
	if info.Length != 0 {
		queryParams.Set("duration", fmt.Sprintf("%.2f", info.Length.Seconds()))
	}
	params := queryParams.Encode()

	url := fmt.Sprintf("%s?%s", LrclibEndpoint, params)
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
		Log(err)
	}

	Log("Fetching lyrics from LRCLIB:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		os.WriteFile(lyricsNotFoundFile, []byte(url), 644)
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
