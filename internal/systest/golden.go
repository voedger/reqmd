package systest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// parseReqGoldenData parses TestMarkdown files to extract golden test data
// It takes a path to the req folder and returns a structured goldenReqData object
func parseReqGoldenData(reqFolderPath string) (*goldenReqData, error) {
	// Initialize the goldenReqData structure
	result := &goldenReqData{
		errors:       make(map[string][]*goldenReqItem),
		reqsites:     make(map[string][]*goldenReqItem),
		footnotes:    make(map[string][]*goldenReqItem),
		newfootnotes: make(map[string][]*goldenReqItem),
	}

	// Walk through the req folder to find TestMarkdown files
	files, err := filepath.Glob(filepath.Join(reqFolderPath, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("error finding TestMarkdown files: %v", err)
	}

	for _, filePath := range files {
		// Read file contents
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		// Process the file line by line
		lines := strings.Split(string(content), "\n")
		previousLineN := 0

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Skip if not a golden line
			if !strings.HasPrefix(trimmedLine, "//") {
				previousLineN = i + 1 // Store current line number for reference
				continue
			}

			// Remove the "//" prefix and trim whitespace
			goldenContent := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "//"))

			// Process errors line
			if strings.HasPrefix(goldenContent, "errors:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("errors line found without preceding content at %s:%d", filePath, i+1)
				}

				// Extract error regexes from the line
				errorPart := strings.TrimSpace(strings.TrimPrefix(goldenContent, "errors:"))
				reErrPattern := regexp.MustCompile(`"([^"]*)"`)
				matches := reErrPattern.FindAllStringSubmatch(errorPart, -1)

				for _, match := range matches {
					pattern := match[1]
					regex, err := regexp.Compile(pattern)
					if err != nil {
						return nil, fmt.Errorf("invalid error regex at %s:%d: %v", filePath, i+1, err)
					}

					item := &goldenReqItem{
						expectedLineN: previousLineN,
						path:          filePath,
						regex:         regex,
					}

					result.errors[filePath] = append(result.errors[filePath], item)
				}
				continue
			}

			// Process reqsite line
			if strings.HasPrefix(goldenContent, "reqsite:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("reqsite line found without preceding content at %s:%d", filePath, i+1)
				}

				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "reqsite:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					expectedLineN: previousLineN,
					path:          filePath,
					data:          data,
				}

				result.reqsites[filePath] = append(result.reqsites[filePath], item)
				continue
			}

			// Process footnote line
			if strings.HasPrefix(goldenContent, "footnote:") {
				if previousLineN == 0 {
					return nil, fmt.Errorf("footnote line found without preceding content at %s:%d", filePath, i+1)
				}

				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "footnote:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					expectedLineN: previousLineN,
					path:          filePath,
					data:          data,
				}

				result.footnotes[filePath] = append(result.footnotes[filePath], item)
				continue
			}

			// Process newfootnote line
			if strings.HasPrefix(goldenContent, "newfootnote:") {
				data := strings.TrimSpace(strings.TrimPrefix(goldenContent, "newfootnote:"))
				// Replace backticks with double quotes
				data = strings.ReplaceAll(data, "`", "\"")

				item := &goldenReqItem{
					expectedLineN: 0, // For newfootnotes, expectedLineN is 0
					path:          filePath,
					data:          data,
				}

				result.newfootnotes[filePath] = append(result.newfootnotes[filePath], item)
				continue
			}
		}
	}

	return result, nil
}

// goldenReqData holds the parsed golden data from TestMarkdown files
type goldenReqData struct {
	errors       map[string][]*goldenReqItem
	reqsites     map[string][]*goldenReqItem
	footnotes    map[string][]*goldenReqItem
	newfootnotes map[string][]*goldenReqItem
}

// goldenReqItem represents a single golden data item with its context
type goldenReqItem struct {
	expectedLineN int            // Line number where the content is expected (0 for newfootnotes)
	path          string         // File path
	data          string         // Content for reqsites, footnotes, and newfootnotes
	regex         *regexp.Regexp // Compiled regex for errors
}
