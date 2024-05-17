package check

import (
	"errors"

	"github.com/judimator/augurken/formatter"
	"github.com/spf13/cobra"
)

func NewCommand(logger formatter.Log) *cobra.Command {
	var indent int
	cmd := &cobra.Command{
		Use:   "check [file or path]",
		Short: "Check formatting of gherkin file(s)",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("please, specify file or folder")
			}

			indent, _ := cmd.Flags().GetInt("indent")
			fileManager := formatter.NewFileManager(indent, logger)
			if err := fileManager.Check(args[0]); err != nil {
				return err
			}

			return nil
		},
	}
	cmd.Flags().IntVarP(&indent, "indent", "i", 2, "set the indentation for Gherkin features (default 2)")
	return cmd
}
