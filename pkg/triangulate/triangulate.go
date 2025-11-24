package triangulate

import (
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
	EnvVarEnable     bool
	EnvVarName       string
	CaseSensitiveSet bool
	MaxDepthSet      bool
	EnvVarEnableSet  bool
	EnvVarNameSet    bool
}

type triangulateConfig struct {
	MarkerFile     string   `json:"marker_file"`
	MarkerFiles    []string `json:"marker_files"`
	StartDirectory string   `json:"start_directory"`
	CaseSensitive  *bool    `json:"case_sensitive"`
	MaxDepth       *int     `json:"max_depth"`
	EnvVarEnable   *bool    `json:"env_var_enable"`
	EnvVarName     string   `json:"env_var_name"`
}

var (
	defaultMarkerFiles   = []string{"TRIANGULATE"}
	defaultCaseSensitive = true
	defaultEnvVarName    = "TRIANGULATE_ROOT"
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

	if opts.EnvVarEnable && strings.TrimSpace(opts.EnvVarName) == "" {
		opts.EnvVarName = defaultEnvVarName
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
	if src.EnvVarEnableSet {
		o.EnvVarEnable = src.EnvVarEnable
		o.EnvVarEnableSet = true
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
	if raw := strings.TrimSpace(os.Getenv("TRIANGULATE_ENV_VAR_ENABLE")); raw != "" {
		if val, err := strconv.ParseBool(raw); err == nil {
			o.EnvVarEnable = val
			o.EnvVarEnableSet = true
		}
	}
	if name := strings.TrimSpace(os.Getenv("TRIANGULATE_ENV_VAR_NAME")); name != "" {
		o.EnvVarName = name
		o.EnvVarNameSet = true
	}
}

func (o *Options) applyConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var section triangulateConfig
	if err := json.Unmarshal(data, &section); err != nil {
		return
	}

	markers := normalizedMarkers(append([]string{}, section.MarkerFiles...))
	if section.MarkerFile != "" {
		markers = append(markers, section.MarkerFile)
	}

	if len(markers) > 0 {
		o.MarkerFiles = markers
	}
	if section.StartDirectory != "" {
		o.StartDir = section.StartDirectory
	}
	if section.CaseSensitive != nil {
		o.CaseSensitive = *section.CaseSensitive
		o.CaseSensitiveSet = true
	}
	if section.MaxDepth != nil && *section.MaxDepth >= 0 {
		o.MaxDepth = *section.MaxDepth
		o.MaxDepthSet = true
	}
	if section.EnvVarEnable != nil {
		o.EnvVarEnable = *section.EnvVarEnable
		o.EnvVarEnableSet = true
	}
	if trimmed := strings.TrimSpace(section.EnvVarName); trimmed != "" {
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
