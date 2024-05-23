package cmd

import (
	"fmt"
	"runtime"

	"github.com/judimator/augurken/cmd/check"
	"github.com/judimator/augurken/cmd/format"
	"github.com/judimator/augurken/meta"
	"github.com/spf13/cobra"
)

func NewCommand(cmdName string) *cobra.Command {
	cmd := &cobra.Command{
		Use: cmdName,
		Version: fmt.Sprintf(
			"%s (build time: %s, %s), OS: %s, arch: %s",
			meta.Version(),
			meta.BuildTime(),
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		),
	}
	cmd.AddCommand(format.NewCommand(), check.NewCommand())

	return cmd
}
