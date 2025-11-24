package main

import (
	"fmt"
	"os"

	"github.com/programmablemike/triangulate/internal/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "triangulate",
		Usage:     "Identify project root directories",
		ArgsUsage: "[start-directory]",
		Flags:     cmd.RootFlags(),
		Action:    cmd.RootAction,
		Commands: []*cli.Command{
			cmd.NewShellCommand(),
			cmd.NewConfigCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "triangulate: %v\n", err)
		os.Exit(1)
	}
}
