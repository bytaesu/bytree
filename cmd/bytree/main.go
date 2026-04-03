package main

import (
	"os"

	"github.com/bytaesu/bytree/internal/cmd"
)

var version = "dev"

func main() {
	root := cmd.NewRoot(version)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
