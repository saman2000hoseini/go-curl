package main

import (
	"os"

	"github.com/saman2000hoseini/go-curl/cmd"
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
