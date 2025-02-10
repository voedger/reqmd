package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// mdParser implements IScanner for markdown files
type mdParser struct{}

// Regular expressions for parsing markdown elements
var (
	headerRegex        = regexp.MustCompile(`^reqmd\.package:\s*(.+)$`)
	requirementRegex   = regexp.MustCompile("`~([^~]+)~`")
	coverageAnnotRegex = regexp.MustCompile(`~([^~]+)~coverage\[\^~[^~]+~\]`)

	// RequirementSite
	RequirementSiteRegex = regexp.MustCompile("`(~[A-Za-z][A-Za-z0-9_]*(?:\\.[A-Za-z][A-Za-z0-9_]*)*~)`(?:cov\\[\\^(~[A-Za-z][A-Za-z0-9_]*(?:\\.[A-Za-z][A-Za-z0-9_]*)*~)\\])?")
)

func NewMarkdownParser() IScanner {
	return &mdParser{}
}

func (m *mdParser) Scan(paths []string) ([]FileStructure, []SyntaxError) {
	var files []FileStructure
	var errors []SyntaxError

	for _, path := range paths {
		// Walk through directory
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip non-markdown files
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(filePath), ".md") {
				file, errs := m.parseFile(filePath)
				if file != nil {
					files = append(files, *file)
				}
				errors = append(errors, errs...)
			}
			return nil
		})

		if err != nil {
			errors = append(errors, SyntaxError{
				FilePath: path,
				Message:  "Failed to walk directory: " + err.Error(),
			})
		}
	}

	return files, errors
}

func (m *mdParser) parseFile(filePath string) (*FileStructure, []SyntaxError) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, []SyntaxError{{
			FilePath: filePath,
			Message:  "Failed to open file: " + err.Error(),
		}}
	}
	defer file.Close()

	var errors []SyntaxError
	structure := &FileStructure{
		Path: filePath,
		Type: FileTypeMarkdown,
	}

	// Parse header and content
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inHeader := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle header section
		if line == "---" {
			if !inHeader {
				inHeader = true
				continue
			} else {
				inHeader = false
				continue
			}
		}

		if inHeader {
			if matches := headerRegex.FindStringSubmatch(line); len(matches) > 1 {
				structure.PackageID = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Parse requirements
		requirements := m.parseRequirements(line, lineNum)
		structure.Requirements = append(structure.Requirements, requirements...)
	}

	if err := scanner.Err(); err != nil {
		errors = append(errors, SyntaxError{
			FilePath: filePath,
			Message:  "Error reading file: " + err.Error(),
		})
	}

	return structure, errors
}

func (m *mdParser) parseRequirements(line string, lineNum int) []Requirement {
	var requirements []Requirement

	// Find all requirement references
	matches := requirementRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 {
			req := Requirement{
				ID:          match[1], // Will be prefixed with PackageID later
				Line:        lineNum,
				IsAnnotated: coverageAnnotRegex.MatchString(line),
			}
			requirements = append(requirements, req)
		}
	}

	return requirements
}
