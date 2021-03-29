package main

import (
	"go-curl/cmd"
	"os"
)

const (
	exitFailure = 1
)

func main() {
	root := cmd.NewCommand()

	if root != nil {
		if err := root.Execute(); err != nil {
			os.Exit(exitFailure)
		}
	}
}
