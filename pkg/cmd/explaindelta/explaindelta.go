package explaindelta

import (
	"errors"
	"github.com/spf13/cobra"
	"os"
)

type ExplainDeltaOptions struct {
	DeltaFile string
}

func NewCmdExplainDelta() *cobra.Command {
	deltaOpts := &ExplainDeltaOptions{}
	cmd := &cobra.Command{
		Use:  "explain-delta <delta-file>",
		Long: "Prints instructions from a delta file; useful when debugging.",
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			argOffset := 0
			if deltaOpts.DeltaFile == "" && len(args) > argOffset {
				deltaOpts.DeltaFile = args[argOffset]
				argOffset += 1
			}
			return explainDeltaRun(deltaOpts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&deltaOpts.DeltaFile, "delta-file", "", "", "The file to explain.")

	return cmd
}

func explainDeltaRun(opts *ExplainDeltaOptions) error {
	deltaFilePath := opts.DeltaFile

	if deltaFilePath == "" {
		return errors.New("no delta file was specified")
	}

	deltaFile, err := os.Open(deltaFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("delta file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = deltaFile.Close() }()
	deltaFileInfo, err := deltaFile.Stat()
	if err != nil {
		return err
	}

	print(deltaFileInfo)
	// reader := octodiff.DeltaReader(deltaFile, deltaFileInfo.Size())
	return nil
}
