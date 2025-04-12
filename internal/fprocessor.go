// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FolderProcessor is a function type that processes a folder and returns a FileProcessor
// for handling files within that folder. If an error occurs during folder processing,
// it returns the error and a nil FileProcessor.
// The folder path provided is always absolute.
type FolderProcessor func(absFolderPath string) (FileProcessor, error)

// FileProcessor is a function type that processes a single file and returns an error
// if the processing fails.
// The file path provided is always absolute.
type FileProcessor func(absFilePath string) error

// FoldersScanner performs concurrent processing of files in a directory tree.
// It traverses the directory structure in breadth-first order, using FolderProcessor
// to obtain FileProcessor instances for each folder, then processes files using
// a pool of goroutines. All paths passed to processors are absolute paths.
//
// Parameters:
//   - nroutines: Number of concurrent goroutines for file processing (must be > 0)
//   - nerrors: Maximum number of errors to buffer in the error channel
//   - root: Root directory path to start scanning from (will be converted to absolute)
//   - fp: FolderProcessor function to process folders and obtain FileProcessors
//
// Returns:
//   - []error: Slice containing all errors encountered during scanning and processing.
//     Returns nil if no errors occurred.
//
// Behavior:
//   - Traverses directories breadth-first to maintain predictable processing order
//   - Processes files concurrently using a worker pool of size nroutines
//   - Collects all errors from both folder and file processing
//   - Stops folder processing on folder processor error but continues with other folders
//   - Continues processing even if some files fail, collecting all errors
func FoldersScanner(nroutines int, nerrors int, root string, fp FolderProcessor) []error {
	if nroutines < 1 {
		return []error{fmt.Errorf("number of routines must be positive")}
	}

	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return []error{fmt.Errorf("failed to get absolute path: %v", err)}
	}

	// Channel for collecting file processors
	fileProcessors := make(chan struct {
		processor FileProcessor
		path      string
	})

	// Channel for collecting errors
	errorsChan := make(chan error, nerrors)
	var errors []error

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < nroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range fileProcessors {
				if err := task.processor(task.path); err != nil {
					select {
					case errorsChan <- err:
					default:
						// Channel is full, log or handle accordingly
					}
				}
			}
		}()
	}

	// Process folders breadth-first
	folders := []string{absRoot}
	for len(folders) > 0 {
		currentFolder := folders[0]
		folders = folders[1:]

		// Get file processor for current folder
		fileProcessor, err := fp(currentFolder)
		if err != nil {
			select {
			case errorsChan <- err:
			default:
				// Channel is full, log or handle accordingly
			}
			continue
		}

		// Skip folder processing if fileProcessor is nil
		if fileProcessor == nil {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(currentFolder)
		if err != nil {
			select {
			case errorsChan <- err:
			default:
				// Channel is full, log or handle accordingly
			}
			continue
		}

		// Process entries
		for _, entry := range entries {
			path := filepath.Join(currentFolder, entry.Name())

			if entry.IsDir() {
				// Add subfolder to the queue
				folders = append(folders, path)
			} else {
				// Send file to processing pool
				fileProcessors <- struct {
					processor FileProcessor
					path      string
				}{fileProcessor, path}
			}
		}
	}

	// Close file processors channel and wait for workers to finish
	close(fileProcessors)
	wg.Wait()

	// Collect all errors
	close(errorsChan)
	for err := range errorsChan {
		errors = append(errors, err)
	}

	return errors
}
