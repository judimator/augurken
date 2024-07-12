package main

import (
	"os"
	"path/filepath"

	"github.com/judimator/augurken/cmd"
)

var exitFn = os.Exit

func main() { exitFn(run()) }

func run() int {
	command := cmd.NewCommand(filepath.Base(os.Args[0]))

	// `err` just helps to guess whether command successful or not. To see proper log result DO NOT log anything here
	if err := command.Execute(); err != nil {
		return 1
	}

	return 0
}
