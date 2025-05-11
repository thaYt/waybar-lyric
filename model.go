package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Nadim147c/go-mpris"
)

// LrcLibResponse is the response sent from LrcLib api
type LrcLibResponse struct {
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

// LyricLine is a line of synchronized lyrics
type LyricLine struct {
	Timestamp time.Duration
	Text      string
}

// Lyrics is a slice of LyricLine
type Lyrics []LyricLine

// Status is the alt/class for waybar
type Status string

const (
	Music   Status = "music"
	Lyric   Status = "lyric"
	Playing Status = "playing"
	Paused  Status = "paused"
)

type Class []Status

type Waybar struct {
	Text    string `json:"text"`
	Class   Class  `json:"class"`
	Alt     Status `json:"alt"`
	Tooltip string `json:"tooltip"`
}

func NewWaybar(lyrics []LyricLine, idx int) *Waybar {
	lyric := lyrics[idx]
	start := max(idx-2, 0)
	end := min(idx+TooltipLines-2, len(lyrics))

	tooltipLyrics := lyrics[start:end]

	var tooltip strings.Builder

	tooltip.WriteString(fmt.Sprintf("<span foreground=\"%s\">", TooltipColor))

	for i, ttl := range tooltipLyrics {
		line := ttl.Text
		if ttl.Text == "" {
			line = "󰝚 "
		}

		if start+i == idx {
			newLine := fmt.Sprintf("</span><b><big>%s</big></b>\n<span foreground=\"%s\">", line, TooltipColor)
			tooltip.WriteString(newLine)
			continue
		}

		tooltip.WriteString(line + "\n")
	}

	line := truncate(lyric.Text)
	tt := strings.TrimSpace(tooltip.String()) + "</span>"

	class := Class{Lyric, Playing}
	return &Waybar{Alt: Lyric, Class: class, Text: line, Tooltip: tt}
}

func (w *Waybar) Encode() {
	if Compact {
		fmt.Println(w.Text)
		return
	}

	if w == (&Waybar{}) {
		fmt.Println("{}")
	}

	e := json.NewEncoder(os.Stdout)
	e.SetEscapeHTML(false)
	e.Encode(w)
}

func (w *Waybar) Is(waybar *Waybar) bool {
	return reflect.DeepEqual(w, waybar)
}

func (w *Waybar) Paused(info *PlayerInfo) {
	if !LyricOnly {
		w.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
	}
	w.Alt = Paused
	w.Class = Class{Paused}
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

func (p *PlayerInfo) Waybar() *Waybar {
	var alt Status = Playing
	if p.Status == "Paused" {
		alt = Paused
	}

	text := fmt.Sprintf("%s - %s", p.Artist, p.Title)

	return &Waybar{Class: Class{alt}, Text: text, Alt: alt}
}

type Store map[string]Lyrics

// Save saves lyrics to Store
func (s Store) Save(key string, value Lyrics) {
	s[key] = value
}

// Load loads lyrics from Store
func (s Store) Load(key string) (Lyrics, bool) {
	v, e := s[key]
	return v, e
}
