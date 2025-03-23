// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func ExecRootCmd(args []string, ver string) error {
	rootCmd := prepareRootCmd(
		"reqmd",
		"Requirements markdown processor",
		args,
		ver,
		newTraceCmd(),
	)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
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

	cmd := &cobra.Command{
		Use:           "trace [flags] <path-to-markdowns> [source-paths...]",
		Short:         "Trace requirements in markdown files",
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			reqPath := args[0]
			srcPaths := args[1:]

			scanner := NewScanner(extensions)
			analyzer := NewAnalyzer()
			applier := NewApplier(dryRun)

			tracer := NewTracer(scanner, analyzer, applier, reqPath, srcPaths)

			return tracer.Trace()
		},
	}

	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Comma-separated list of source file extensions to process (e.g. .go,.ts,.js)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be done, but make no changes to files")
	return cmd
}
