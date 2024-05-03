package format

import (
	"errors"

	"github.com/judimator/augurken/formatter"
	"github.com/spf13/cobra"
)

func NewCommand(logger formatter.Log) *cobra.Command {
	var indent uint8
	cmd := &cobra.Command{
		Use:   "format",
		Short: "format gherkin file(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("please, specify file or folder")
			}

			fileManager := formatter.NewFileManager(int(indent), logger)
			if err := fileManager.FormatAndReplace(args[0]); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().Uint8VarP(&indent, "indent", "i", 2, "set the indentation for Gherkin steps, examples")
	return cmd
}
