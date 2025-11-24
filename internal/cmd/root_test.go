package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestRootActionUsesPositionalStartDir(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(root, "TRIANGULATE"), "")

	buf := &bytes.Buffer{}
	app := &cli.App{
		Writer: buf,
		Flags:  RootFlags(),
		Action: RootAction,
	}

	origExiter := cli.OsExiter
	cli.OsExiter = func(int) {}
	defer func() { cli.OsExiter = origExiter }()

	if err := app.Run([]string{"triangulate", deep}); err != nil {
		t.Fatalf("app.Run: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != root {
		t.Fatalf("stdout = %q; want %q", got, root)
	}
}

func TestRootActionRejectsStartFlagWithPositional(t *testing.T) {
	app := &cli.App{
		Flags:  RootFlags(),
		Action: RootAction,
	}

	origExiter := cli.OsExiter
	cli.OsExiter = func(int) {}
	defer func() { cli.OsExiter = origExiter }()

	err := app.Run([]string{"triangulate", "--start", "/one", "/two"})
	if err == nil {
		t.Fatalf("app.Run succeeded; expected error")
	}

	if exitErr, ok := err.(cli.ExitCoder); !ok || exitErr.ExitCode() == 0 {
		t.Fatalf("unexpected error type or code: %v", err)
	}
}

func TestRootActionSetsEnvVarDefaultName(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(root, "TRIANGULATE"), "")

	const defaultName = "TRIANGULATE_ROOT"
	t.Setenv(defaultName, "")

	app := &cli.App{
		Flags:  RootFlags(),
		Action: RootAction,
	}

	origExiter := cli.OsExiter
	cli.OsExiter = func(int) {}
	defer func() { cli.OsExiter = origExiter }()

	if err := app.Run([]string{"triangulate", "--env-var-enable", deep}); err != nil {
		t.Fatalf("app.Run: %v", err)
	}

	if got := os.Getenv(defaultName); got != root {
		t.Fatalf("%s = %q; want %q", defaultName, got, root)
	}
}

func TestRootActionSetsEnvVarCustomName(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b")
	mustMkdirAll(t, deep)
	mustWriteFile(t, filepath.Join(root, "TRIANGULATE"), "")

	const envName = "MY_PROJECT_ROOT"
	t.Setenv(envName, "")

	app := &cli.App{
		Flags:  RootFlags(),
		Action: RootAction,
	}

	origExiter := cli.OsExiter
	cli.OsExiter = func(int) {}
	defer func() { cli.OsExiter = origExiter }()

	if err := app.Run([]string{"triangulate", "--env-var-enable", "--env-var-name", envName, deep}); err != nil {
		t.Fatalf("app.Run: %v", err)
	}

	if got := os.Getenv(envName); got != root {
		t.Fatalf("%s = %q; want %q", envName, got, root)
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
