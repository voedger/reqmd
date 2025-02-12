package internal

/*

An exerpt from design.md

- **Purpose**: Implements `IScanner`.
- **Key functions**:
  - `Scan`:
    - Recursively discover Markdown and source files.
    - Delegate parsing to specialized components (`mdparser.go`, `srccoverparser.go`).
    - Build a unified list of `FileStructure` objects for each file.
    - Collect any `SyntaxError`s.
- **Responsibilities**:
  - Single responsibility: collecting raw data (files, coverage tags, requirement references) and building the domain model.
  - Potential concurrency (goroutines) for scanning subfolders.

*/

type ScanResult struct {
	Files        []FileStructure
	SyntaxErrors []SyntaxError
}

/*

- Paths are processes sequentially by FoldersScanner using 32 routines
- First path is processed as path to requirement files
- Other paths are processed as path to source files using

Requirement files
- Uses FoldersScanner and ParseMarkdownFile
- FolderProcessor parses reqmdfiles.json (if exists) and passes to FileProcessor

Source files
- git repository shall be found in Path or parent directories
- IGit is instantiated using NewIGit and a folder that contains .git folder
- IGit should be passed to FolderProcessor and all FileProcessors
- Uses FoldersScanner and ParseSourceFile

*/

func Scan(paths []string) (res *ScanResult, err error) {
	return nil, nil
}
