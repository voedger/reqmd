package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/voedger/reqmd/internal"
)

//go:embed version
var version string

var verbose bool

func main() {
	if err := execRootCmd(os.Args, version); err != nil {
		os.Exit(1)
	}
}

func execRootCmd(args []string, ver string) error {
	rootCmd := prepareRootCmd(
		"reqmd",
		"Requirements markdown processor",
		args,
		ver,
		newTraceCmd(),
	)

	return rootCmd.Execute()
}

func prepareRootCmd(use, short string, args []string, ver string, cmds ...*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     use,
		Version: ver,
		Short:   short,
	}
	rootCmd.SetArgs(args[1:])
	rootCmd.AddCommand(cmds...)
	return rootCmd
}

func newTraceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace <path-to-markdowns> [source-paths...]",
		Short: "Trace requirements in markdown files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reqPath := args[0]
			srcPaths := args[1:]

			scanner := internal.NewScanner()
			analyzer := internal.NewAnalyzer()
			applier := internal.NewDummyApplier()

			tracer := internal.NewTracer(scanner, analyzer, applier, reqPath, srcPaths)

			if err := tracer.Trace(); err != nil {
				return fmt.Errorf("error: %v", err)
			}

			if verbose {
				fmt.Println("Processing completed successfully")
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output showing detailed processing information")
	return cmd
}
