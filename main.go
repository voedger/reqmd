package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/voedger/reqmd/internal"
)

func main() {
	// Set up the trace subcommand
	traceCmd := flag.NewFlagSet("trace", flag.ExitOnError)
	verbose := traceCmd.Bool("v", false, "Enable verbose output showing detailed processing information")

	// Validate command
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Expected 'trace' subcommand\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "trace":
		handleTrace(traceCmd, os.Args[2:], verbose)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", os.Args[1])
		os.Exit(1)
	}
}

func handleTrace(traceCmd *flag.FlagSet, args []string, verbose *bool) {
	err := traceCmd.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing trace command: %v\n", err)
		os.Exit(1)
	}

	// Validate required path argument
	if traceCmd.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Required <path-to-markdowns> argument missing\n")
		os.Exit(1)
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

	// Execute trace operation and handle exit codes per requirements
	if err := tracer.Trace(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1) // Syntax errors during scan
	}

	// Success
	if *verbose {
		fmt.Println("Processing completed successfully")
	}
}
