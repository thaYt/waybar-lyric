package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nadim147c/go-mpris"
)

type (
	LrcLibResponse struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		TrackName    string  `json:"trackName"`
		ArtistName   string  `json:"artistName"`
		AlbumName    string  `json:"albumName"`
		Duration     float64 `json:"duration"`
		Instrumental bool    `json:"instrumental"`
		PlainLyrics  string  `json:"plainLyrics"`
		SyncedLyrics string  `json:"syncedLyrics"`
	}

	LyricLine struct {
		Timestamp time.Duration
		Text      string
	}
)

type Status string

const (
	Music   Status = "music"
	Lyric   Status = "lyric"
	Playing Status = "playing"
	Paused  Status = "paused"
)

type Class []Status

type Waybar struct {
	Text       string `json:"text"`
	Class      Class  `json:"class"`
	Alt        Status `json:"alt"`
	Tooltip    string `json:"tooltip"`
	Percentage int    `json:"percentage"`
}

func NewWaybar(lyrics []LyricLine, idx, percentage, maxLineLength int) *Waybar {
	lyric := lyrics[idx]
	start := max(idx-2, 0)
	end := min(idx+5, len(lyrics))

	tooltipLyrics := lyrics[start:end]
	var tooltip strings.Builder
	for i, ttl := range tooltipLyrics {
		if start+i == idx {
			tooltip.WriteString("> ")
		}
		if ttl.Text != "" {
			tooltip.WriteString(ttl.Text + "\n")
		} else {
			tooltip.WriteString("󰝚 \n")
		}

	}

	line := truncate(lyric.Text, maxLineLength)
	tt := strings.TrimSpace(tooltip.String())

	return &Waybar{
		Alt: Lyric, Class: Class{Lyric, Playing},
		Text:       line,
		Tooltip:    tt,
		Percentage: percentage,
	}
}

func (w *Waybar) Encode() {
	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.Encode(w)
}

type PlayerInfo struct {
	ID     string
	Artist string
	Title  string
	Album  string

	Position time.Duration
	Length   time.Duration

	Status mpris.PlaybackStatus
}

func (p *PlayerInfo) Percentage() int {
	return int((p.Position * 100) / p.Length)
}

func (p *PlayerInfo) Waybar() *Waybar {
	var alt Status = Playing
	if p.Status == "Paused" {
		alt = Paused
	}

	text := fmt.Sprintf("%s - %s", p.Artist, p.Title)

	return &Waybar{
		Class:      Class{alt},
		Text:       text,
		Alt:        alt,
		Percentage: p.Percentage(),
	}
}
