package main

import (
	"os"

	"github.com/hmschreck/netbeat/cmd"

	// _ "github.com/hmschreck/netbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
