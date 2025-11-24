package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/programmablemike/triangulate/pkg/triangulate"
	"github.com/urfave/cli/v2"
)

func NewConfigCommand() *cli.Command {
	configFlag := &cli.StringFlag{
		Name:  "config",
		Value: triangulate.DefaultConfigPath(),
		Usage: "Path to config file",
	}

	return &cli.Command{
		Name:  "config",
		Usage: "Manage triangulate configuration",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all configuration values",
				Flags:  []cli.Flag{configFlag},
				Action: configList,
			},
			{
				Name:      "get",
				Usage:     "Get a configuration value",
				ArgsUsage: "<key>",
				Flags:     []cli.Flag{configFlag},
				Action:    configGet,
			},
			{
				Name:      "set",
				Usage:     "Set a configuration value",
				ArgsUsage: "<key> <value>",
				Flags:     []cli.Flag{configFlag},
				Action:    configSet,
			},
			{
				Name:   "validate",
				Usage:  "Validate configuration file format",
				Flags:  []cli.Flag{configFlag},
				Action: configValidate,
			},
		},
	}
}

func configList(c *cli.Context) error {
	cfg, err := triangulate.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}

	data, err := jsonPretty(trimConfig(cfg))
	if err != nil {
		return err
	}
	fmt.Fprintln(c.App.Writer, string(data))
	return nil
}

func configGet(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.Exit("usage: triangulate config get <key>", 1)
	}
	key := strings.ToLower(c.Args().First())

	cfg, err := triangulate.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}

	val, err := lookupConfigValue(cfg, key)
	if err != nil {
		return err
	}
	fmt.Fprintln(c.App.Writer, val)
	return nil
}

func configSet(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.Exit("usage: triangulate config set <key> <value>", 1)
	}
	key := strings.ToLower(c.Args().First())
	value := c.Args().Get(1)

	cfg, err := triangulate.ReadConfig(c.String("config"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		cfg = triangulate.Config{}
	}

	if err := setConfigValue(&cfg, key, value); err != nil {
		return err
	}

	if err := triangulate.WriteConfig(c.String("config"), cfg); err != nil {
		return err
	}
	return nil
}

func configValidate(c *cli.Context) error {
	if err := triangulate.ValidateConfigFile(c.String("config")); err != nil {
		return err
	}
	fmt.Fprintln(c.App.Writer, "valid")
	return nil
}

func lookupConfigValue(cfg triangulate.Config, key string) (string, error) {
	switch key {
	case "marker_file":
		if cfg.MarkerFile == "" {
			return "", errors.New("marker_file not set")
		}
		return cfg.MarkerFile, nil
	case "marker_files":
		if len(cfg.MarkerFiles) == 0 {
			return "", errors.New("marker_files not set")
		}
		return strings.Join(cfg.MarkerFiles, ","), nil
	case "start_directory":
		if cfg.StartDirectory == "" {
			return "", errors.New("start_directory not set")
		}
		return cfg.StartDirectory, nil
	case "case_sensitive":
		if cfg.CaseSensitive == nil {
			return "", errors.New("case_sensitive not set")
		}
		return strconv.FormatBool(*cfg.CaseSensitive), nil
	case "max_depth":
		if cfg.MaxDepth == nil {
			return "", errors.New("max_depth not set")
		}
		return strconv.Itoa(*cfg.MaxDepth), nil
	case "env_var_name":
		if cfg.EnvVarName == "" {
			return "", errors.New("env_var_name not set")
		}
		return cfg.EnvVarName, nil
	default:
		return "", fmt.Errorf("unknown key %q", key)
	}
}

func setConfigValue(cfg *triangulate.Config, key, value string) error {
	switch key {
	case "marker_file":
		cfg.MarkerFile = strings.TrimSpace(value)
	case "marker_files":
		cfg.MarkerFiles = splitComma(value)
	case "start_directory":
		cfg.StartDirectory = strings.TrimSpace(value)
	case "case_sensitive":
		val, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("parse bool: %w", err)
		}
		cfg.CaseSensitive = &val
	case "max_depth":
		val, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("parse int: %w", err)
		}
		if val < 0 {
			return fmt.Errorf("max_depth must be >= 0")
		}
		cfg.MaxDepth = &val
	case "env_var_name":
		cfg.EnvVarName = strings.TrimSpace(value)
	default:
		return fmt.Errorf("unknown key %q", key)
	}
	return nil
}

func jsonPretty(cfg triangulate.Config) ([]byte, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func trimConfig(cfg triangulate.Config) triangulate.Config {
	out := triangulate.Config{}
	if cfg.MarkerFile != "" {
		out.MarkerFile = cfg.MarkerFile
	}
	if len(cfg.MarkerFiles) > 0 {
		out.MarkerFiles = append([]string{}, cfg.MarkerFiles...)
	}
	if cfg.StartDirectory != "" {
		out.StartDirectory = cfg.StartDirectory
	}
	if cfg.CaseSensitive != nil {
		out.CaseSensitive = cfg.CaseSensitive
	}
	if cfg.MaxDepth != nil {
		out.MaxDepth = cfg.MaxDepth
	}
	if cfg.EnvVarName != "" {
		out.EnvVarName = cfg.EnvVarName
	}
	return out
}
