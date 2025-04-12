// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

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
	paths    []string // For multi-path approach
}

// NewTracer creates a tracer that handles multiple paths for both markdown and source files
func NewTracer(scanner IScanner, analyzer IAnalyzer, applier IApplier, paths []string) ITracer {
	return &tracer{
		scanner:  scanner,
		analyzer: analyzer,
		applier:  applier,
		paths:    paths,
	}
}

func (t *tracer) Trace() error {
	return t.trace()
}

// traceMultiPath handles the new unified approach where multiple paths can contain both markdown and source files
func (t *tracer) trace() error {
	// Get current dir
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	Verbose("Starting processing with multi-path approach", "wd", wd, "paths", fmt.Sprintf("%v", t.paths))

	// Convert to absolute paths
	absolutePaths := make([]string, len(t.paths))
	for i, path := range t.paths {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
		absolutePaths[i] = absolutePath
		Verbose("Absolute path: " + absolutePath)
	}

	// Pass all paths to scanner
	scanResult, err := t.scanner.Scan(absolutePaths)
	if err != nil {
		return err
	}
	if len(scanResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: scanResult.ProcessingErrors}
	}

	// Analyzing phase (same as before)
	analyzeResult, err := t.analyzer.Analyze(scanResult.Files)
	if err != nil {
		return err
	}
	if len(analyzeResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: analyzeResult.ProcessingErrors}
	}

	// Applying phase (same as before)
	if err := t.applier.Apply(analyzeResult); err != nil {
		return err
	}

	return nil
}
