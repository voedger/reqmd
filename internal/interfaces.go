// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

// ITracer defines the high-level interface for tracing workflow.
// It orchestrates scanning, analyzing, and applying changes.
type ITracer interface {
	Trace() error
}

// IScanner is responsible for scanning file paths and parsing them into FileStructures.
type IScanner interface {
	// ScanMultiPath scans multiple paths that can each contain both markdown and source files
	Scan(paths []string) (*ScannerResult, error)
}

// IAnalyzer checks for semantic issues (e.g., unique RequirementIds) and generates Actions.
type IAnalyzer interface {
	Analyze(files []FileStructure) (*AnalyzerResult, error)
}

// IApplier applies the Actions (file updates, footnote generation, etc.).
type IApplier interface {
	Apply(*AnalyzerResult) error
}

type IVCS interface {
	// Slashed, absolute path to the root of the git repository
	PathToRoot() string // TODO: do we need this?
	FileHash(absoluteFilePath string) (relPath, hash string, err error)
	RepoRootFolderURL() string
}
