package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

/*

An exerpt from design.md

### **tracer.go**

- **Purpose**: Implements `ITracer`. This is the **facade** that orchestrates the scanning, analyzing, and applying phases.
- **Key functions**:
  - `Trace()
- **Responsibilities**:
  - High-level workflow control.
  - Enforce the steps: if syntax errors exist, abort; if semantic errors exist, abort; otherwise apply actions.
  - Use injected interfaced (ref. interfaces.go) IScanner, IAnalyzer, IApplier to scan, analyze, and apply changes.

*/

type tracer struct {
	scanner  IScanner
	analyzer IAnalyzer
	applier  IApplier
	reqPath  string
	srcPaths []string
}

func NewTracer(scanner IScanner, analyzer IAnalyzer, applier IApplier, reqPath string, srcPaths []string) ITracer {
	return &tracer{
		scanner:  scanner,
		analyzer: analyzer,
		applier:  applier,
		reqPath:  reqPath,
		srcPaths: srcPaths,
	}
}

func (t *tracer) Trace() error {
	// Make paths absolute
	{

		// Get current dir
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		Verbose("Starting processing", "wd", wd, "reqPath", t.reqPath, "srcPaths", fmt.Sprintf("%v", t.srcPaths))

		t.reqPath, err = filepath.Abs(t.reqPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for requirement path: %w", err)
		}
		Verbose("Absolute requirement path: " + t.reqPath)

		for i, srcPath := range t.srcPaths {
			t.srcPaths[i], err = filepath.Abs(srcPath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for source path: %w", err)
			}
			Verbose("Absolute source path: " + t.srcPaths[i])
		}
	}

	// Scanning phase
	scanResult, err := t.scanner.Scan(t.reqPath, t.srcPaths)
	if err != nil {
		return err
	}
	if len(scanResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: scanResult.ProcessingErrors}
	}

	// Analyzing phase
	analyzeResult, err := t.analyzer.Analyze(scanResult.Files)
	if err != nil {
		return err
	}
	if len(analyzeResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: analyzeResult.ProcessingErrors}
	}

	// Applying phase
	if err := t.applier.Apply(analyzeResult); err != nil {
		return err
	}

	return nil
}
