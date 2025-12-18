package main

import (
	"log"

	"github.com/jcleira/workspace/cmd"
	_ "github.com/jcleira/workspace/cmd/branch"
	_ "github.com/jcleira/workspace/cmd/config"
	_ "github.com/jcleira/workspace/cmd/create"
	_ "github.com/jcleira/workspace/cmd/delete"
	_ "github.com/jcleira/workspace/cmd/list"
	_ "github.com/jcleira/workspace/cmd/switch"
)

func main() {
	if err := cmd.InitializeConfig(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
