package format

import (
	"errors"

	"github.com/judimator/augurken/formatter"
	"github.com/judimator/augurken/log"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var indent int
	cmd := &cobra.Command{
		Use:   "format [file or path]",
		Short: "Format gherkin file(s)",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				err := errors.New("please, specify file or folder")
				log.Error(err)

				return err
			}

			success := true
			indent, _ := cmd.Flags().GetInt("indent")
			fileManager := formatter.NewFileManager(indent)
			result := fileManager.FormatAndReplace(args[0])

			for _, r := range result {
				if s, ok := r.(string); ok {
					log.Success(s)

					continue
				}
				if e, ok := r.(error); ok {
					log.Error(e)
					success = false
				}
			}

			if !success {
				return errors.New("error occurred while formatting file/folder")
			}

			return nil
		},
	}
	cmd.Flags().IntVarP(&indent, "indent", "i", 2, "set the indentation for Gherkin features (default 2)")

	return cmd
}
