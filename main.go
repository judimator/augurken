package main

import (
	"os"
	"path/filepath"

	"github.com/judimator/augurken/cmd"
	"github.com/judimator/augurken/log"
)

var exitFn = os.Exit

func main() { exitFn(run()) }

func run() int {
	command := cmd.NewCommand(filepath.Base(os.Args[0]))

	if err := command.Execute(); err != nil {
		log.Error(err)

		return 1
	}

	return 0
}
