package internal

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type applier struct {
	dryRun bool
}

func NewApplier(dryRun bool) IApplier {
	return &applier{
		dryRun: dryRun,
	}
}

func (a *applier) Apply(ar *AnalyzerResult) error {
	if a.dryRun || IsVerbose {
		Verbose("Actions that would be applied:")
		for _, actions := range ar.MdActions {
			for _, action := range actions {
				Verbose("Action\n\t" + action.String())
			}
		}
		if a.dryRun {
			return nil
		}
	}
	if a.dryRun {
		return nil
	}

	for path, actions := range ar.MdActions {
		err := applyMdActions(path, actions)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Principles:

- RequirementSiteRegex and CoverageFootnoteRegex from models.go are used to match lines with RequirementID
-

*/

func applyMdActions(path FilePath, actions []MdAction) error {
	lines, hasCRLF, err := readFilePreserveEndings(string(path))
	if err != nil {
		return err
	}

	for _, action := range actions {
		if action.Line > 0 {
			lineIndex := action.Line - 1
			if lineIndex < 0 || lineIndex >= len(lines) {
				return fmt.Errorf("line %d doesn't exist in file %s", action.Line, path)
			}
			line := lines[lineIndex]
			var re *regexp.Regexp
			switch action.Type {
			case ActionSite:
				re = RequirementSiteRegex
			case ActionFootnote:
				re = CoverageFootnoteRegex
			default:
				return fmt.Errorf("unknown action type: %s", action.Type)
			}
			if !re.MatchString(line) {
				return fmt.Errorf("line %d does not match requirement ID in file %s", action.Line, path)
			}
			newLine := re.ReplaceAllStringFunc(line, func(_ string) string {
				return action.Data
			})
			lines[lineIndex] = newLine

		} else {
			if action.Type != ActionFootnote {
				return fmt.Errorf("invalid action type for line=0 in file %s", path)
			}
			if needFootnoteSeparator(lines) {
				lines = append(lines, "")
			}
			lines = append(lines, action.Data)
		}
	}

	// if last line is not empty, add an empty line
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}

	if err := writeFilePreserveEndings(string(path), lines, hasCRLF); err != nil {
		return err
	}
	return nil
}

// readFilePreserveEndings reads a file, detects if it uses CRLF, and returns lines without stripping end-of-line markers.
func readFilePreserveEndings(filePath string) ([]string, bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}
	hasCRLF := bytes.Contains(content, []byte("\r\n"))

	var lines []string
	if hasCRLF {
		lines = strings.Split(string(content), "\r\n")
	} else {
		lines = strings.Split(string(content), "\n")
	}
	return lines, hasCRLF, nil
}

// writeFilePreserveEndings joins lines with CRLF or LF depending on hasCRLF and writes them back to disk.
func writeFilePreserveEndings(filePath string, lines []string, hasCRLF bool) error {
	delim := "\n"
	if hasCRLF {
		delim = "\r\n"
	}
	out := strings.Join(lines, delim)
	return os.WriteFile(filePath, []byte(out), 0644)
}

// needFootnoteSeparator checks if we must insert an empty line before the first appended footnote.
func needFootnoteSeparator(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	// If the file already ends with an empty line, no need to add another.
	lastLine := lines[len(lines)-1]
	if strings.TrimSpace(lastLine) == "" {
		return false
	}
	// Check for existing footnotes.
	for _, ln := range lines {
		if CoverageFootnoteRegex.MatchString(ln) {
			return false
		}
	}
	return true
}
