// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"runtime/debug"

	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed version
var Version string

func ExecRootCmd(args []string, ver string) error {
	rootCmd := prepareRootCmd(
		"reqmd",
		"Requirements processor",
		args,
		ver,
		newTraceCmd(),
		newVersionCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}

func newVersionCmd() *cobra.Command {

	info, _ := debug.ReadBuildInfo()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print reqmd version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("info.Main.Version: %s\n", info.Main.Version)
		},
	}
	return cmd

}

func prepareRootCmd(use, short string, args []string, ver string, cmds ...*cobra.Command) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     use,
		Version: ver,
		Short:   short,
	}
	rootCmd.PersistentFlags().BoolVarP(&IsVerbose, "verbose", "v", false, "Enable verbose output showing detailed processing information")
	rootCmd.SetArgs(args[1:])
	rootCmd.AddCommand(cmds...)
	return rootCmd
}

func newTraceCmd() *cobra.Command {
	var extensions string
	var dryRun bool
	var ignoreLines []string

	cmd := &cobra.Command{
		Use:           "trace [flags] <paths>...",
		Short:         "Trace requirements in markdown files",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args

			// Validate all paths exist
			for _, path := range paths {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return fmt.Errorf("path does not exist: %s", path)
				}
			}

			patterns, err := preparePatterns(ignoreLines)
			if err != nil {
				return err
			}
			scanner := NewScanner(extensions, patterns)
			analyzer := NewAnalyzer()
			applier := NewApplier(dryRun)

			tracer := NewTracer(scanner, analyzer, applier, paths)

			return tracer.Trace()
		},
	}

	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Comma-separated list of source file extensions to process (e.g. .go,.ts,.js)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done, but make no changes to files")
	cmd.Flags().StringArrayVar(&ignoreLines, "ignore-lines", nil, "Regular expression pattern for lines to ignore. Can be specified multiple times.")
	return cmd
}
