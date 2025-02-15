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

var (
	extensions string
)

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
	rootCmd.PersistentFlags().BoolVarP(&internal.IsVerbose, "verbose", "v", false, "Enable verbose output showing detailed processing information")
	rootCmd.SetArgs(args[1:])
	rootCmd.AddCommand(cmds...)
	return rootCmd
}

func newTraceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "trace [-e extensions] <path-to-markdowns> [source-paths...]",
		Short:        "Trace requirements in markdown files",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			reqPath := args[0]
			srcPaths := args[1:]

			scanner := internal.NewScanner(extensions)
			analyzer := internal.NewAnalyzer()
			applier := internal.NewDummyApplier()

			internal.Verbose("Starting processing", "reqPath", reqPath, "srcPaths", fmt.Sprintf("%v", srcPaths))

			tracer := internal.NewTracer(scanner, analyzer, applier, reqPath, srcPaths)

			if err := tracer.Trace(); err != nil {
				return fmt.Errorf("error: %v", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Comma-separated list of source file extensions to process (e.g., .go,.ts,.js)")
	return cmd
}
