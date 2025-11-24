package triangulate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Options captures configuration for root discovery and optional override markers.
type Options struct {
	MarkerFiles      []string
	StartDir         string
	CaseSensitive    bool
	MaxDepth         int
	ConfigPath       string
	EnvVarName       string
	CaseSensitiveSet bool
	MaxDepthSet      bool
	EnvVarNameSet    bool
}

// Config represents the persisted configuration file structure.
type Config struct {
	MarkerFile     string   `json:"marker_file,omitempty"`
	MarkerFiles    []string `json:"marker_files,omitempty"`
	StartDirectory string   `json:"start_directory,omitempty"`
	CaseSensitive  *bool    `json:"case_sensitive,omitempty"`
	MaxDepth       *int     `json:"max_depth,omitempty"`
	EnvVarName     string   `json:"env_var_name,omitempty"`
}

var (
	defaultMarkerFiles   = []string{"TRIANGULATE"}
	defaultCaseSensitive = true
	defaultConfigName    = ".triangulate"

	// ErrNotFound indicates no marker file could be found.
	ErrNotFound = errors.New("triangulate: marker not found")
)

// ResolveOptions merges defaults with config, environment variables, and explicit overrides.
// Precedence: defaults < config file < environment variables < explicit Options fields.
func ResolveOptions(src Options) (Options, error) {
	opts := Options{
		MarkerFiles:   append([]string{}, defaultMarkerFiles...),
		CaseSensitive: defaultCaseSensitive,
		ConfigPath:    DefaultConfigPath(),
	}

	if src.ConfigPath != "" {
		opts.ConfigPath = src.ConfigPath
	}

	opts.applyConfig(opts.ConfigPath)
	opts.applyEnv()
	opts.applyOverrides(src)

	if opts.StartDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return Options{}, fmt.Errorf("get working directory: %w", err)
		}
		opts.StartDir = cwd
	}

	if len(opts.MarkerFiles) == 0 {
		opts.MarkerFiles = append([]string{}, defaultMarkerFiles...)
	}

	return opts, nil
}

// FindRoot walks upward from opts.StartDir looking for any configured marker file.
func FindRoot(opts Options) (string, error) {
	start, err := filepath.Abs(opts.StartDir)
	if err != nil {
		return "", fmt.Errorf("abs start dir: %w", err)
	}

	current := start
	for depth := 0; ; depth++ {
		if opts.MaxDepth > 0 && depth > opts.MaxDepth {
			break
		}

		found, err := hasAnyMarker(current, opts.MarkerFiles, opts.CaseSensitive)
		if err != nil {
			return "", err
		}
		if found {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", ErrNotFound
}

func hasAnyMarker(dir string, markers []string, caseSensitive bool) (bool, error) {
	if caseSensitive {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return true, nil
			} else if !errors.Is(err, os.ErrNotExist) {
				return false, fmt.Errorf("stat marker %q: %w", marker, err)
			}
		}
		return false, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		for _, marker := range markers {
			if strings.EqualFold(entry.Name(), marker) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (o *Options) applyOverrides(src Options) {
	if len(src.MarkerFiles) > 0 {
		o.MarkerFiles = normalizedMarkers(src.MarkerFiles)
	}
	if src.StartDir != "" {
		o.StartDir = src.StartDir
	}
	if src.CaseSensitiveSet {
		o.CaseSensitive = src.CaseSensitive
		o.CaseSensitiveSet = true
	}
	if src.MaxDepthSet && src.MaxDepth >= 0 {
		o.MaxDepth = src.MaxDepth
		o.MaxDepthSet = true
	}
	if src.EnvVarNameSet && strings.TrimSpace(src.EnvVarName) != "" {
		o.EnvVarName = strings.TrimSpace(src.EnvVarName)
		o.EnvVarNameSet = true
	}
}

func (o *Options) applyEnv() {
	if markers := strings.TrimSpace(os.Getenv("TRIANGULATE_MARKER_FILE")); markers != "" {
		o.MarkerFiles = normalizedMarkers(strings.Split(markers, ","))
	}
	if start := strings.TrimSpace(os.Getenv("TRIANGULATE_START_DIRECTORY")); start != "" {
		o.StartDir = start
	}
	if raw := strings.TrimSpace(os.Getenv("TRIANGULATE_CASE_SENSITIVE")); raw != "" {
		if val, err := strconv.ParseBool(raw); err == nil {
			o.CaseSensitive = val
			o.CaseSensitiveSet = true
		}
	}
	if raw := strings.TrimSpace(os.Getenv("TRIANGULATE_MAX_DEPTH")); raw != "" {
		if val, err := strconv.Atoi(raw); err == nil && val >= 0 {
			o.MaxDepth = val
			o.MaxDepthSet = true
		}
	}
	if name := strings.TrimSpace(os.Getenv("TRIANGULATE_ENV_VAR_NAME")); name != "" {
		o.EnvVarName = name
		o.EnvVarNameSet = true
	}
}

func (o *Options) applyConfig(path string) {
	cfg, err := ReadConfig(path)
	if err != nil {
		return
	}

	markers := normalizedMarkers(append([]string{}, cfg.MarkerFiles...))
	if cfg.MarkerFile != "" {
		markers = append(markers, cfg.MarkerFile)
	}

	if len(markers) > 0 {
		o.MarkerFiles = markers
	}
	if cfg.StartDirectory != "" {
		o.StartDir = cfg.StartDirectory
	}
	if cfg.CaseSensitive != nil {
		o.CaseSensitive = *cfg.CaseSensitive
		o.CaseSensitiveSet = true
	}
	if cfg.MaxDepth != nil && *cfg.MaxDepth >= 0 {
		o.MaxDepth = *cfg.MaxDepth
		o.MaxDepthSet = true
	}
	if trimmed := strings.TrimSpace(cfg.EnvVarName); trimmed != "" {
		o.EnvVarName = trimmed
		o.EnvVarNameSet = true
	}
}

func normalizedMarkers(markers []string) []string {
	out := make([]string, 0, len(markers))
	for _, m := range markers {
		if trimmed := strings.TrimSpace(m); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// DefaultConfigPath returns the default path to the configuration file.
// It prefers $HOME/.triangulate but falls back to the filename in the cwd if $HOME is unavailable.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return defaultConfigName
	}
	return filepath.Join(home, defaultConfigName)
}

// ParseConfig parses raw JSON configuration data into a Config.
func ParseConfig(data []byte) (Config, error) {
	var cfg Config
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// ReadConfig reads and parses the configuration file from path.
func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return ParseConfig(data)
}

// WriteConfig writes cfg to path in JSON format, creating parent directories if needed.
func WriteConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	return os.WriteFile(path, data, 0o644)
}

// ValidateConfig ensures the provided JSON data conforms to Config.
func ValidateConfig(data []byte) error {
	_, err := ParseConfig(data)
	return err
}

// ValidateConfigFile validates the configuration file at path.
func ValidateConfigFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return ValidateConfig(data)
}
