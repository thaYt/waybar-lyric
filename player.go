package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"

	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
)

type PlayerParser func(*mpris.Player) (*PlayerInfo, error)

var supportedPlayers = map[string]PlayerParser{
	"spotify": DefaultParser,
	"amarok":  DefaultParser,
}

func SelectPlayer(conn *dbus.Conn) (*mpris.Player, PlayerParser, error) {
	players, err := mpris.List(conn)
	if err != nil {
		return nil, nil, err
	}

	if len(players) == 0 {
		return nil, nil, errors.New("No player exists")
	}

	for name, parser := range supportedPlayers {
		for player := range slices.Values(players) {
			if mpris.BaseInterface+"."+name == player {
				return mpris.New(conn, player), parser, nil
			}
		}
	}

	playerName := players[0]
	return mpris.New(conn, playerName), DefaultParser, nil
}

// StringToMD5 converts a string to its MD5 hash
func StringToMD5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// DefaultParser takes *mpris.Player of spotify and return *PlayerInfo
func DefaultParser(player *mpris.Player) (*PlayerInfo, error) {
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
