# waybar-lyric

A CLI tool that displays Spotify lyrics on your
[Waybar](https://github.com/Alexays/Waybar) for Linux systems.

## Description

`waybar-lyric` fetches and displays real-time lyrics on your Waybar. It provides a
scrolling lyrics display that syncs with your currently playing music, enhancing your
desktop music experience.

## Features

- Real-time display of the current song's lyrics
- Click to toggle play/pause
- Smart caching system:
  - Stores available lyrics locally to reduce API requests
  - Remembers songs without lyrics to prevent unnecessary API calls
- Single instance enforcement via file locking (flocking)
- Configurable maximum text length
- Detailed logging options

## Installation

### Prerequisites

- [Waybar](https://github.com/Alexays/Waybar)
- A working Spotify installation
- DBus connectivity

### Install

#### Prerequisites

- [go](https://go.dev/)

#### Installation

- From [AUR](https://aur.archlinux.org/packages)

```bash
yay -S waybar-lyric-git
```

- With `go install`

```bash
go install github.com/Nadim147c/waybar-lyric@latest
```

- Or install from source

```bash
git clone https://github.com/Nadim147c/waybar-lyric.git
cd waybar-lyric
go install
```

## Usage

```
Usage: /usr/bin/waybar-lyric [options]
Get spotify lyrics on waybar.

Options:
  --init              Show json snippet for waybar/config.jsonc
  --log-file string   File to where logs should be save
  --max-length int    Maximum lenght of lyrics text (default 150)
  --toggle            Toggle player state (pause/resume)
  -v, --verbose       Use verbose loggin
  --version           Print the version of waybar-lyric
```

## Configuration

### Waybar Configuration

The recommended way to configure waybar-lyric is to generate the configuration
snippet using the built-in command:

```bash
waybar-lyric --init
```

This will output the proper JSON configuration snippet that you can copy directly
into your Waybar `config.jsonc` file.

### Style Example

Add to your `style.css`:

```css
#custom-spotify {
  color: #1db954;
  margin: 0 5px;
  padding: 0 10px;
}
```

## Troubleshooting

If you encounter issues:

1. Check that Spotify is running and connected
2. Run with verbose logging: `waybar-lyric -v --log-file=/tmp/waybar-lyric.log`
3. Verify DBus connectivity with: `dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:PlaybackStatus`

## License

This repository is licensed under [AGPL-3.0](./LICENSE). Thanks to
[LrcLib](https://lrclib.net/) for providing lyrics.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
