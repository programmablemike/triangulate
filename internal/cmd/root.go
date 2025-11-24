package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/programmablemike/triangulate/pkg/triangulate"
	"github.com/urfave/cli/v2"
)

// RootFlags defines the flags available on the default command.
func RootFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "marker",
			Usage: "Marker file name(s), comma-separated",
		},
		&cli.StringFlag{
			Name:  "start",
			Usage: "Start directory for search",
		},
		&cli.BoolFlag{
			Name:  "case-sensitive",
			Usage: "Set true/false for case sensitivity",
		},
		&cli.IntFlag{
			Name:  "max-depth",
			Value: -1,
			Usage: "Maximum directory levels to search (0 for current only)",
		},
		&cli.StringFlag{
			Name:  "config",
			Value: triangulate.DefaultConfigPath(),
			Usage: "Path to config file",
		},
		&cli.StringFlag{
			Name:  "env-var-name",
			Usage: "Name of environment variable to set with the triangulated path",
		},
	}
}

func RootAction(c *cli.Context) error {
	if c.NArg() > 1 {
		return cli.Exit("usage: triangulate [start-directory]", 1)
	}

	startArg := c.Args().First()
	src := triangulate.Options{
		ConfigPath: c.String("config"),
	}

	if markers := splitComma(c.String("marker")); len(markers) > 0 {
		src.MarkerFiles = markers
	}
	if startArg != "" && c.IsSet("start") {
		return cli.Exit("cannot use --start with positional start-directory", 1)
	}
	if startArg != "" {
		src.StartDir = startArg
	} else if start := c.String("start"); start != "" {
		src.StartDir = start
	}
	if c.IsSet("case-sensitive") {
		val := c.Bool("case-sensitive")
		src.CaseSensitive = val
		src.CaseSensitiveSet = true
	}
	if md := c.Int("max-depth"); md >= 0 {
		src.MaxDepth = md
		src.MaxDepthSet = true
	}
	if name := strings.TrimSpace(c.String("env-var-name")); name != "" {
		src.EnvVarName = name
		src.EnvVarNameSet = true
	}

	opts, err := triangulate.ResolveOptions(src)
	if err != nil {
		return err
	}

	root, err := triangulate.FindRoot(opts)
	if err != nil {
		return err
	}

	if opts.EnvVarName != "" {
		name := opts.EnvVarName
		if err := os.Setenv(name, root); err != nil {
			return fmt.Errorf("set env var %q: %w", name, err)
		}
	}

	fmt.Fprintln(c.App.Writer, root)
	return nil
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
