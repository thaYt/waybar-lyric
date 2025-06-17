package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	NoLyric Status = "no_lyric"
)

type Class []Status

type Waybar struct {
	Text    string      `json:"text"`
	Class   Class       `json:"class"`
	Alt     Status      `json:"alt"`
	Tooltip string      `json:"tooltip"`
	Info    *PlayerInfo `json:"info,omitempty"`
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
	if LyricOnly && (w.Alt == Paused || w.Alt == Music) {
		fmt.Println("")
		return
	}

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

// PlayerInfo holds all information of currently playing track metadata
type PlayerInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Album  string `json:"album"`

	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`

	Status mpris.PlaybackStatus `json:"status"`
}

func (p *PlayerInfo) Waybar() *Waybar {
	var alt Status = Playing
	if p.Status == "Paused" {
		alt = Paused
	}

	var text string
	if !LyricOnly {
		text = fmt.Sprintf("%s - %s", p.Artist, p.Title)
	}

	waybar := &Waybar{
		Class: Class{alt},
		Text:  text,
		Alt:   alt,
	}
	if Detailed {
		waybar.Info = p
	}
	return waybar
}

// UpdatePosition updates the position of player
func (p *PlayerInfo) UpdatePosition(player *mpris.Player) error {
	pos, err := player.GetPosition()
	if err != nil {
		return err
	}
	p.Position = pos

	// HACK: YoutubeMusic dbus position ≈ 1.1 slow
	if player.GetName() == mpris.BaseInterface+".YoutubeMusic" {
		slog.Debug("Adding 1.1 second to adjust mpris delay")
		p.Position += 1100 * time.Millisecond
	}

	return nil
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

// noopHandler is a empty handler that ignore all logs
type noopHandler struct{}

var _ slog.Handler = (*noopHandler)(nil)

func (h *noopHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (h *noopHandler) Handle(_ context.Context, _ slog.Record) error { return nil }
func (h *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return h }
func (h *noopHandler) WithGroup(_ string) slog.Handler               { return h }
