//go:build ignore

package main

import (
	"os"
	"path/filepath"

	"github.com/jcleira/workspace/cmd"
	_ "github.com/jcleira/workspace/cmd/branch"
	_ "github.com/jcleira/workspace/cmd/config"
	_ "github.com/jcleira/workspace/cmd/create"
	_ "github.com/jcleira/workspace/cmd/delete"
	_ "github.com/jcleira/workspace/cmd/list"
	_ "github.com/jcleira/workspace/cmd/switch"
)

func main() {
	dir := "completions"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	bashFile, err := os.Create(filepath.Join(dir, "workspace.bash"))
	if err != nil {
		panic(err)
	}
	defer func() { _ = bashFile.Close() }()
	if err := cmd.RootCmd.GenBashCompletion(bashFile); err != nil {
		panic(err)
	}

	zshFile, err := os.Create(filepath.Join(dir, "workspace.zsh"))
	if err != nil {
		panic(err)
	}
	defer func() { _ = zshFile.Close() }()
	if err := cmd.RootCmd.GenZshCompletion(zshFile); err != nil {
		panic(err)
	}

	fishFile, err := os.Create(filepath.Join(dir, "workspace.fish"))
	if err != nil {
		panic(err)
	}
	defer func() { _ = fishFile.Close() }()
	if err := cmd.RootCmd.GenFishCompletion(fishFile, true); err != nil {
		panic(err)
	}
}
