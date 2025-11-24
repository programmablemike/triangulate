package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func NewShellCommand() *cli.Command {
	return &cli.Command{
		Name:   "shell",
		Usage:  "Emit shell integration snippet",
		Action: shellAction,
	}
}

func shellAction(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.Exit("usage: triangulate shell <bash|zsh>", 1)
	}

	switch c.Args().First() {
	case "bash":
		fmt.Fprint(c.App.Writer, bashSnippet())
	case "zsh":
		fmt.Fprint(c.App.Writer, zshSnippet())
	default:
		return cli.Exit(fmt.Sprintf("unsupported shell %q (use bash or zsh)", c.Args().First()), 1)
	}

	return nil
}

func bashSnippet() string {
	return `# triangulate shell integration
triangulate__set_root() {
  local root
  root=$(command triangulate "$@" 2>/dev/null) || root=""
  local var_name=${TRIANGULATE_ENV_VAR_NAME:-TRIANGULATE_ROOT}
  if [ -n "$root" ]; then
    export "${var_name}=${root}"
  else
    unset "$var_name"
  fi
}

triangulate__last_pwd=""
triangulate__maybe_refresh() {
  if [ "$PWD" != "$triangulate__last_pwd" ]; then
    triangulate__last_pwd="$PWD"
    triangulate__set_root
  fi
}

triangulate__maybe_refresh

if [ -n "${PROMPT_COMMAND:+x}" ]; then
  PROMPT_COMMAND="triangulate__maybe_refresh;${PROMPT_COMMAND}"
else
  PROMPT_COMMAND="triangulate__maybe_refresh"
fi
`
}

func zshSnippet() string {
	return `# triangulate shell integration
triangulate__set_root() {
  local root
  root=$(command triangulate "$@" 2>/dev/null) || root=""
  local var_name=${TRIANGULATE_ENV_VAR_NAME:-TRIANGULATE_ROOT}
  if [ -n "$root" ]; then
    export "${var_name}=${root}"
  else
    unset "$var_name"
  fi
}

triangulate__last_pwd=""
triangulate__maybe_refresh() {
  if [ "$PWD" != "$triangulate__last_pwd" ]; then
    triangulate__last_pwd="$PWD"
    triangulate__set_root
  fi
}

triangulate__maybe_refresh

autoload -Uz add-zsh-hook 2>/dev/null
if command -v add-zsh-hook >/dev/null 2>&1; then
  add-zsh-hook chpwd triangulate__maybe_refresh
fi
`
}
