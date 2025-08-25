package main

import (
	"os"

	"github.com/Nadim147c/waybar-lyric/cmd"
)

func main() {
	if err := cmd.Command.Execute(); err != nil {
		os.Exit(1)
	}
}
