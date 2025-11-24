package triangulate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRootCaseSensitive(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b", "c")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(dir, "TRIANGULATE"), "")

	opts, err := ResolveOptions(Options{
		StartDir:   deep,
		ConfigPath: filepath.Join(dir, ".triangulate"),
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	got, err := FindRoot(opts)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != dir {
		t.Fatalf("FindRoot = %q; want %q", got, dir)
	}
}

func TestFindRootCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(dir, "triangulate"), "")

	opts, err := ResolveOptions(Options{
		StartDir:         deep,
		ConfigPath:       filepath.Join(dir, ".triangulate"),
		CaseSensitive:    false,
		CaseSensitiveSet: true,
		MarkerFiles:      []string{"TRIANGULATE"},
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	got, err := FindRoot(opts)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != dir {
		t.Fatalf("FindRoot = %q; want %q", got, dir)
	}
}

func TestResolveOptionsPrecedence(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".triangulate")
	mustWriteFile(t, configPath, `
{
  "marker_file":"CONFIG_MARKER",
  "start_directory":"/config",
  "case_sensitive":true,
  "max_depth":1,
  "env_var_name":"CONFIG_VAR"
}
`)

	t.Setenv("TRIANGULATE_MARKER_FILE", "ENV_MARKER")
	t.Setenv("TRIANGULATE_START_DIRECTORY", "/env")
	t.Setenv("TRIANGULATE_CASE_SENSITIVE", "false")
	t.Setenv("TRIANGULATE_MAX_DEPTH", "3")
	t.Setenv("TRIANGULATE_ENV_VAR_ENABLE", "false")
	t.Setenv("TRIANGULATE_ENV_VAR_NAME", "ENV_VAR")

	opts, err := ResolveOptions(Options{
		ConfigPath:       configPath,
		MarkerFiles:      []string{"CLI_MARKER"},
		StartDir:         "/cli",
		CaseSensitive:    true,
		CaseSensitiveSet: true,
		MaxDepth:         5,
		MaxDepthSet:      true,
		EnvVarName:       "CLI_VAR",
		EnvVarNameSet:    true,
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	wantMarkers := []string{"CLI_MARKER"}
	if len(opts.MarkerFiles) != 1 || opts.MarkerFiles[0] != wantMarkers[0] {
		t.Fatalf("MarkerFiles = %v; want %v", opts.MarkerFiles, wantMarkers)
	}
	if opts.StartDir != "/cli" {
		t.Fatalf("StartDir = %q; want /cli", opts.StartDir)
	}
	if opts.CaseSensitive != true {
		t.Fatalf("CaseSensitive = %v; want true", opts.CaseSensitive)
	}
	if opts.MaxDepth != 5 {
		t.Fatalf("MaxDepth = %d; want 5", opts.MaxDepth)
	}
	if opts.EnvVarName != "CLI_VAR" {
		t.Fatalf("EnvVarName = %q; want CLI_VAR", opts.EnvVarName)
	}
}

func TestResolveOptionsEnvOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".triangulate")
	mustWriteFile(t, configPath, `
{
  "marker_file":"CONFIG_MARKER",
  "start_directory":"/config",
  "case_sensitive":true,
  "max_depth":4,
  "env_var_name":"CONFIG_VAR"
}
`)

	t.Setenv("TRIANGULATE_MARKER_FILE", "ENV_MARKER")
	t.Setenv("TRIANGULATE_START_DIRECTORY", "/env")
	t.Setenv("TRIANGULATE_CASE_SENSITIVE", "false")
	t.Setenv("TRIANGULATE_MAX_DEPTH", "2")
	t.Setenv("TRIANGULATE_ENV_VAR_NAME", "ENV_VAR")

	opts, err := ResolveOptions(Options{
		ConfigPath: configPath,
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	if len(opts.MarkerFiles) != 1 || opts.MarkerFiles[0] != "ENV_MARKER" {
		t.Fatalf("MarkerFiles = %v; want [ENV_MARKER]", opts.MarkerFiles)
	}
	if opts.StartDir != "/env" {
		t.Fatalf("StartDir = %q; want /env", opts.StartDir)
	}
	if opts.CaseSensitive {
		t.Fatalf("CaseSensitive = %v; want false", opts.CaseSensitive)
	}
	if opts.MaxDepth != 2 {
		t.Fatalf("MaxDepth = %d; want 2", opts.MaxDepth)
	}
	if opts.EnvVarName != "ENV_VAR" {
		t.Fatalf("EnvVarName = %q; want ENV_VAR", opts.EnvVarName)
	}
}

func TestFindRootMaxDepth(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "TRIANGULATE"), "")
	deep := filepath.Join(dir, "a", "b", "c", "d")
	mustMkdirAll(t, deep)

	opts, err := ResolveOptions(Options{
		StartDir:    deep,
		ConfigPath:  filepath.Join(dir, ".triangulate"),
		MaxDepth:    1,
		MaxDepthSet: true,
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	if _, err := FindRoot(opts); err == nil {
		t.Fatalf("FindRoot succeeded; expected error due to max depth")
	}
}

func TestFindRootPrefersNearestMarker(t *testing.T) {
	root := t.TempDir()
	mid := filepath.Join(root, "mid")
	deep := filepath.Join(mid, "a", "b")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(root, "TRIANGULATE"), "")
	mustWriteFile(t, filepath.Join(mid, "TRIANGULATE"), "")

	opts, err := ResolveOptions(Options{
		StartDir:   deep,
		ConfigPath: filepath.Join(root, ".triangulate"),
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	got, err := FindRoot(opts)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if got != mid {
		t.Fatalf("FindRoot = %q; want %q", got, mid)
	}
}

func TestResolveOptionsEnvironment(t *testing.T) {
	dir := t.TempDir()

	t.Setenv("TRIANGULATE_MARKER_FILE", "ENV_MARKER_ONE, ENV_MARKER_TWO")
	t.Setenv("TRIANGULATE_START_DIRECTORY", dir)
	t.Setenv("TRIANGULATE_CASE_SENSITIVE", "false")
	t.Setenv("TRIANGULATE_MAX_DEPTH", "2")
	t.Setenv("TRIANGULATE_ENV_VAR_NAME", "ENV_ONLY_VAR")

	opts, err := ResolveOptions(Options{
		ConfigPath: filepath.Join(dir, ".triangulate"),
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	wantMarkers := []string{"ENV_MARKER_ONE", "ENV_MARKER_TWO"}
	if len(opts.MarkerFiles) != len(wantMarkers) {
		t.Fatalf("MarkerFiles length = %d; want %d", len(opts.MarkerFiles), len(wantMarkers))
	}
	for i, want := range wantMarkers {
		if opts.MarkerFiles[i] != want {
			t.Fatalf("MarkerFiles[%d] = %q; want %q", i, opts.MarkerFiles[i], want)
		}
	}
	if opts.StartDir != dir {
		t.Fatalf("StartDir = %q; want %q", opts.StartDir, dir)
	}
	if opts.CaseSensitive {
		t.Fatalf("CaseSensitive = %v; want false", opts.CaseSensitive)
	}
	if opts.MaxDepth != 2 {
		t.Fatalf("MaxDepth = %d; want 2", opts.MaxDepth)
	}
	if opts.EnvVarName != "ENV_ONLY_VAR" {
		t.Fatalf("EnvVarName = %q; want ENV_ONLY_VAR", opts.EnvVarName)
	}
}

func TestResolveOptionsTopLevelConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".triangulate")
	mustWriteFile(t, configPath, `
{
  "marker_files": ["TOP_MARKER"],
  "env_var_name": "TOP_VAR"
}
`)

	opts, err := ResolveOptions(Options{
		ConfigPath: configPath,
	})
	if err != nil {
		t.Fatalf("ResolveOptions: %v", err)
	}

	if len(opts.MarkerFiles) != 1 || opts.MarkerFiles[0] != "TOP_MARKER" {
		t.Fatalf("MarkerFiles = %v; want [TOP_MARKER]", opts.MarkerFiles)
	}
	if opts.EnvVarName != "TOP_VAR" {
		t.Fatalf("EnvVarName = %q; want TOP_VAR", opts.EnvVarName)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
