# triangulate

Triangulate is a command-line tool for identifying a project's root directory
path.

## installation

### Homebrew

```shell
brew tap programmablemike/homebrew-tap
brew install triangulate
```

## license

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.

## description

Triangulate is a tool for triangulating the root directory of a project. It does
this by recursively searching the path of the current working directory until it
finds a marker file.

## examples

By default we search for the root directory marked by the file `TRIANGULATE`
starting at the current working directory.

For the directory structure:

```text
/home
  /mike
    /myproject
      TRIANGULATE
      go.pkg
      main.go
      /lib
        /logging
          logging.go
        /cmd
          cmd.go
        /server
          server.go
```

We run the following commands:

```shell
# Navigate to a directory in the project
$: cd /home/mike/lib/logging
# Run a command using triangulate to set the current directory to the root directory first
$: pushd $(triangulate); go build -o ./output/myproject; popd;

# You can also start from an explicit directory path (positional arg)
$: triangulate /home/mike/lib/logging
/home/mike/myproject
```

For flexibility we allow overriding the default file and starting path using
environment variables (`TRIANGULATE_START_DIRECTORY` and
`TRIANGULATE_MARKER_FILE`).

You can also set the value through the `.triangulate` configuration file (by default
loaded from `$HOME/.triangulate`; override with `--config`). Settings live at the top level.

```json
{
  "marker_file": "BUILD.root",
  "start_directory": "/home/mike/myproject",
  "case_sensitive": true,
  "max_depth": 5
}
```

Triangulate is case-sensitive by default, but this can be overridden through the configuration.

## setting an environment variable

One of the main use cases for Triangulate is to be able to keep an environment variable
up-to-date with the project root directory automatically as you navigate in a shell.

This behavior is enabled automatically when `--env-var-name` (or `TRIANGULATE_ENV_VAR_NAME`)
is set. By default we set `TRIANGULATE_ROOT`; override the name with `--env-var-name` or
`TRIANGULATE_ENV_VAR_NAME`:

```shell
$ triangulate --env-var-name PROJECT_ROOT
/path/to/project
$ echo "$PROJECT_ROOT"
/path/to/project
```

## using the SDK

If you need to add triangulate functionality to a Go project you can use the
Triangulate SDK which is defined in the `/pkg` directory.

To import it:

```go
package main

import (
    "fmt"
    "os"
    "github.com/programmablemike/triangulate"
)

func main() {
    opts, err := triangulate.ResolveOptions(triangulate.Options{})
    if err != nil {
        panic(err)
    }
    root, err := triangulate.FindRoot(opts)
    if err != nil {
        panic(err)
    }
    fmt.Println("project root= ", root)
}
```

## installing the shell extension

Triangulate works as a shell extension for `bash` or `zsh`. To install run the shell command
with the name of the shell you are using. This will output a command you can add directory to
your shell configuration `.zshrc` or `.bashrc`.

You can also directly source it using `eval` if you want to avoid polluting your shell configuration with code

```shell
$: triangulate shell zsh >> .zshrc
# or source directly
$: eval $(triangulate shell zsh)
```

The shell snippet installs a hook that re-runs `triangulate` whenever `$PWD` changes and keeps
`$TRIANGULATE_ROOT` (or `TRIANGULATE_ENV_VAR_NAME` if set) up-to-date with the detected project
root.

## environment variables

Here's a list of environment variables you can set.

`TRIANGULATE_MARKER_FILE`: Set the file that marks the root directory.

`TRIANGULATE_START_DIRECTORY`: Set the default directory.

`TRIANGULATE_CASE_SENSITIVE`: Set the case sensitivity.

`TRIANGULATE_MAX_DEPTH`: Set the maximum search depth.

`TRIANGULATE_ENV_VAR_NAME`: The name of the environment variable to set. When set, the value is exported automatically.

## configuration example

Here's the configuration values that can be set.

```json
{
  "marker_files": ["BUILD.root"],
  "start_directory": "/home/mike/myproject",
  "case_sensitive": true,
  "max_depth": 10,
  "env_var_name": "TRIANGULATE_PROJECT_ROOT"
}
```

Manage the file from the CLI with `triangulate config`: use `list` to view all settings,
`get <key>` / `set <key> <value>` to query or update fields, and `validate` to check format.
