package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/Nadim147c/go-mpris"
)

// StringToMD5 converts a string to its MD5 hash
func StringToMD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GetSpotifyInfo takes *mpris.Player of spotify and return *PlayerInfo
func GetSpotifyInfo(player *mpris.Player) (*PlayerInfo, error) {
	meta, err := player.GetMetadata()
	if err != nil {
		return nil, err
	}

	status, err := player.GetPlaybackStatus()
	if err != nil {
		return nil, err
	}

	position, err := player.GetPosition()
	if err != nil {
		return nil, err
	}

	artistList, ok := meta["xesam:artist"].Value().([]string)
	if !ok || len(artistList) == 0 {
		return nil, fmt.Errorf("missing artist information")
	}
	artist := artistList[0]

	title, ok := meta["xesam:title"].Value().(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("missing title information")
	}

	id, ok := meta["mpris:trackid"].Value().(string)
	if !ok || id == "" {
		id = StringToMD5(artist + title)
	}

	album, _ := meta["xesam:album"].Value().(string)
	length, err := player.GetLength()
	if err != nil {
		return nil, err
	}

	return &PlayerInfo{
		ID:       id,
		Artist:   artist,
		Title:    title,
		Album:    album,
		Status:   status,
		Position: position,
		Length:   length,
	}, nil
}
