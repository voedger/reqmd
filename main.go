package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/voedger/reqmd/internal"
)

//go:embed version
var version string

func main() {
	if err := execRootCmd(os.Args, version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execRootCmd(args []string, _ string) error {
	// Set up the trace subcommand
	traceCmd := flag.NewFlagSet("trace", flag.ExitOnError)
	verbose := traceCmd.Bool("v", false, "Enable verbose output showing detailed processing information")


	// Validate command
	if len(args) < 2 {
		return fmt.Errorf("Expected 'trace' subcommand")
	}

	switch args[1] {
	case "trace":
		return handleTrace(traceCmd, args[2:], verbose)
	default:
		return fmt.Errorf("Unknown command %q", args[1])
	}
}

func handleTrace(traceCmd *flag.FlagSet, args []string, verbose *bool) error {
	err := traceCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing trace command: %v", err)
	}

	// Validate required path argument
	if traceCmd.NArg() < 1 {
		return fmt.Errorf("required <path-to-markdowns> argument missing")
	}

	// Get paths
	reqPath := traceCmd.Arg(0)      // First arg is requirements path
	srcPaths := traceCmd.Args()[1:] // Remaining args are source paths

	// Initialize components
	scanner := internal.NewScanner()
	analyzer := internal.NewAnalyzer()
	applier := internal.NewDummyApplier()

	// Create and run tracer
	tracer := internal.NewTracer(scanner, analyzer, applier, reqPath, srcPaths)

	// Execute trace operation
	if err := tracer.Trace(); err != nil {
		return fmt.Errorf("error: %v", err)
	}

	// Success
	if *verbose {
		fmt.Println("Processing completed successfully")
	}

	return nil
}
