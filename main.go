package main

import (
	"os"
	"path/filepath"

	"github.com/judimator/augurken/cmd"
	"github.com/judimator/augurken/formatter"
)

var exitFn = os.Exit

func main() { exitFn(run()) }

func run() int {
	logger := formatter.NewLogger()
	command := cmd.NewCommand(filepath.Base(os.Args[0]), logger)

	if err := command.Execute(); err != nil {
		logger.Error(err)

		return 1
	}

	return 0
}
