package cmd

import (
	"strings"
	"testing"
)

func TestBashSnippetContainsHooks(t *testing.T) {
	snippet := bashSnippet()
	for _, want := range []string{
		"triangulate__maybe_refresh",
		"PROMPT_COMMAND",
		"TRIANGULATE_ENV_VAR_NAME:-TRIANGULATE_ROOT",
	} {
		if !strings.Contains(snippet, want) {
			t.Fatalf("bash snippet missing %q", want)
		}
	}
}

func TestZshSnippetContainsHooks(t *testing.T) {
	snippet := zshSnippet()
	for _, want := range []string{
		"add-zsh-hook chpwd triangulate__maybe_refresh",
		"triangulate__maybe_refresh",
		"TRIANGULATE_ENV_VAR_NAME:-TRIANGULATE_ROOT",
	} {
		if !strings.Contains(snippet, want) {
			t.Fatalf("zsh snippet missing %q", want)
		}
	}
}
