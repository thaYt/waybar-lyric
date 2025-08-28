package init

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed init.jsonc
var snippet string

var nobat bool

func init() {
	// This flag for ignoring by the init command when running for deprecreated
	// waybar-lyric --init
	Command.Flags().BoolP("init", "i", false, "Toggle player state between pause and resume")
	Command.Flags().BoolVarP(&nobat, "no-bat", "n", false, "Avoid using bat for syntext highlighting")
	Command.Flags().MarkHidden("init")
}

// Command is the play-pause command
var Command = &cobra.Command{
	Use:          "init",
	Short:        "Print json snippet for waybar confg",
	SilenceUsage: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Fprint(os.Stderr, "Put the following object in your waybar config:\n\n")

		if !nobat {
			err := exec.Command("which", "bat").Run()
			if err != nil {
				fmt.Println(snippet)
				return err
			}

			batCmd := exec.Command("bat", "-pljson")

			reader := strings.NewReader(snippet)
			batCmd.Stdin = reader
			batCmd.Stdout = os.Stdout

			if err := batCmd.Run(); err != nil {
				fmt.Println(snippet)
				return err
			}

			return nil
		}

		fmt.Println(snippet)
		return nil
	},
}
