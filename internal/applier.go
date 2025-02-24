package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	for path, actions := range ar.MdActions {
		err := a.applyMdActions(path, actions)
		if err != nil {
			return err
		}
	}
	for path, reqmdjson := range ar.Reqmdjsons {
		err := a.applyReqmdjson(path, reqmdjson)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Principles:

- RequirementSiteRegex and CoverageFootnoteRegex from models.go are used to match lines with RequirementId

*/

func (a *applier) applyMdActions(path FilePath, actions []MdAction) error {
	// Read file with preserved line endings
	lines, hasCRLF, err := readFilePreserveEndings(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			lines = []string{}
		} else {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
	}

	// First validate all actions
	for _, action := range actions {
		if action.Line > 0 {
			if action.Line > len(lines) {
				return fmt.Errorf("line %d does not exist in file %s", action.Line, path)
			}

			line := lines[action.Line-1]
			switch action.Type {
			case ActionSite:
				if !RequirementSiteRegex.MatchString(line) ||
					!strings.Contains(line, string(action.RequirementName)) {
					return fmt.Errorf("line %d in file %s does not contain valid requirement site for %s",
						action.Line, path, action.RequirementName)
				}
			case ActionFootnote:
				if !CoverageFootnoteRegex.MatchString(line) ||
					!strings.Contains(line, string(action.RequirementName)) {
					return fmt.Errorf("line %d in file %s does not contain valid footnote for %s",
						action.Line, path, action.RequirementName)
				}
			}
		}
	}

	// Apply actions
	for _, action := range actions {
		if action.Line > 0 {
			a.logOrVerbose("Action\n\t" + action.String())
			// Update existing line
			line := lines[action.Line-1]
			switch action.Type {
			case ActionSite:
				lines[action.Line-1] = RequirementSiteRegex.ReplaceAllString(line, action.Data)
			case ActionFootnote:
				lines[action.Line-1] = CoverageFootnoteRegex.ReplaceAllString(line, action.Data)
			}
		}
	}

	// Process trailing empty lines and footnotes
	// First trim trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Add empty line before footnotes if needed
	hasNewFootnotes := false
	for _, action := range actions {
		if action.Line == 0 && action.Type == ActionFootnote {
			hasNewFootnotes = true
			break
		}
	}

	if hasNewFootnotes && len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if !CoverageFootnoteRegex.MatchString(lastLine) {
			lines = append(lines, "")
		}
	}

	// Append new footnotes
	for _, action := range actions {
		if action.Line == 0 && action.Type == ActionFootnote {
			a.logOrVerbose("Action\n\t" + action.String())
			lines = append(lines, action.Data)
		}
	}

	// Markdown files should end with a newline
	lines = append(lines, "")

	// Write file if not in dry run mode
	a.logOrVerbose("Update markdown file", "path", path)

	if !a.dryRun {
		if err := writeFilePreserveEndings(string(path), lines, hasCRLF); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

func (a *applier) applyReqmdjson(folder_ FolderPath, reqmdjson *Reqmdjson) error {
	filePath := filepath.Join(string(folder_), ReqmdjsonFileName)
	filePath = filepath.ToSlash(filePath)
	if len(reqmdjson.FileUrl2FileHash) == 0 {
		// If reqmdjson is empty and file exists, delete it
		if _, err := os.Stat(string(filePath)); err == nil {
			a.logOrVerbose("Delete reqmd.json ", "path", filePath)
			if !a.dryRun {
				if err := os.Remove(string(filePath)); err != nil {
					return fmt.Errorf("failed to delete reqmd.json at %s: %w", filePath, err)
				}
			}
		} else {
			a.logOrVerbose(ReqmdjsonFileName+" needs to be emptied, but it does not exist yet", "path", filePath)
		}
		return nil
	}

	// Marshal using custom MarshalJSON that ensures ordered FileURLs
	data, err := json.MarshalIndent(reqmdjson, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reqmdjson for %s: %w", folder_, err)
	}

	a.logOrVerbose("Write reqmd.json", "path", filePath, "data", string(data))
	// Write to file
	if !a.dryRun {
		if err := os.WriteFile(string(filePath), data, 0644); err != nil {
			return fmt.Errorf("failed to write reqmd.json to %s: %w", filePath, err)
		}
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

func (a *applier) logOrVerbose(msg string, kv ...any) {
	if a.dryRun || IsVerbose {
		if a.dryRun {
			msg = "[DRY RUN] " + msg
		}
		Log(msg, kv...)
	}
}
