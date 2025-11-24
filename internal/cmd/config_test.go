package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/programmablemike/triangulate/pkg/triangulate"
	"github.com/urfave/cli/v2"
)

func TestConfigList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".triangulate")
	cfg := triangulate.Config{
		MarkerFile:     "BUILD.root",
		MarkerFiles:    []string{"A", "B"},
		StartDirectory: "/tmp/project",
	}
	if err := triangulate.WriteConfig(cfgPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	buf := &bytes.Buffer{}
	app := &cli.App{
		Writer:   buf,
		Commands: []*cli.Command{NewConfigCommand()},
	}

	overrideExit(t)
	if err := app.Run([]string{"triangulate", "config", "list"}); err != nil {
		t.Fatalf("app.Run: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"marker_file": "BUILD.root"`) {
		t.Fatalf("list output missing marker_file: %s", out)
	}
	if strings.Contains(out, `"case_sensitive"`) {
		t.Fatalf("expected unset fields omitted, got: %s", out)
	}
}

func TestConfigGetAndSet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".triangulate")

	app := &cli.App{
		Commands: []*cli.Command{NewConfigCommand()},
	}

	overrideExit(t)
	if err := app.Run([]string{"triangulate", "config", "set", "marker_file", "CONFIG_MARKER"}); err != nil {
		t.Fatalf("set marker_file: %v", err)
	}
	if err := app.Run([]string{"triangulate", "config", "set", "max_depth", "3"}); err != nil {
		t.Fatalf("set max_depth: %v", err)
	}

	cfg, err := triangulate.ReadConfig(cfgPath)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if cfg.MarkerFile != "CONFIG_MARKER" {
		t.Fatalf("MarkerFile = %q; want CONFIG_MARKER", cfg.MarkerFile)
	}
	if cfg.MaxDepth == nil || *cfg.MaxDepth != 3 {
		t.Fatalf("MaxDepth = %v; want 3", cfg.MaxDepth)
	}

	buf := &bytes.Buffer{}
	app.Writer = buf
	if err := app.Run([]string{"triangulate", "config", "get", "marker_file"}); err != nil {
		t.Fatalf("get marker_file: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "CONFIG_MARKER" {
		t.Fatalf("get marker_file output = %q; want CONFIG_MARKER", buf.String())
	}
}

func TestConfigValidate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(home, ".triangulate")
	cfg := triangulate.Config{MarkerFile: "BUILD.root"}
	if err := triangulate.WriteConfig(cfgPath, cfg); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	app := &cli.App{
		Commands: []*cli.Command{NewConfigCommand()},
	}

	overrideExit(t)
	if err := app.Run([]string{"triangulate", "config", "validate"}); err != nil {
		t.Fatalf("validate: %v", err)
	}

	// Corrupt the file with unknown fields.
	if err := os.WriteFile(cfgPath, []byte(`{"unknown": true}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := app.Run([]string{"triangulate", "config", "validate"}); err == nil {
		t.Fatalf("validate succeeded; expected failure")
	}
}

func overrideExit(t *testing.T) {
	t.Helper()
	origExiter := cli.OsExiter
	cli.OsExiter = func(int) {}
	t.Cleanup(func() {
		cli.OsExiter = origExiter
	})
}
