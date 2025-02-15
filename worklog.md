# Worklog

This file contains a brief log of the project design and construction.

## First attempts

**Attempt 1**. ChatGPT o3-mini-high. [private chat](https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed)

- Prompt: Suggest architecture of the solution (Go). Do not generate all code yet. Provide list of files, key functions, structs and their resposibilities.
  - After the analysys of the output the Construction requirements section was added.

**Attempt 2**. ChatGPT o3-mini-high. [private chat](https://chatgpt.com/c/67a90782-3644-800d-a619-956119cc2b0c)

- Propose a solution architecture using SOLID principles.. Don't generate all the code yet. Provide a list of files, key functions, structures, and their responsibilities.
- Generate internal/models.go
- Generate internal/interfaces.go
  - Fix package names
- Note: FileStructure was not defined
- ChatGPT o1:
  - Prompt: Propose a solution architecture using SOLID principles. Don't generate any code yet. Provide a list of files, key functions, structures, and their responsibilities.
  - Generate models.go and interfaces.go
- GitHub Copilot: Suggest mdparser implementation ❌
- Copilot.Claude: Implement mdparser.go ✅
- Copilot.Claude: Generate tests and testdata for mdparser.go
- ChatGPT o1: Generate engaging README.md for this project ❌
- ChatGPT 4o: Generate engaging README.md for this project ✅
- NI: Improve mdparser_test.go
- Copilot.Claude: Generate definition for requirementSiteRegex ❌
- ChatGPT o3-mini-high: Generate definition (Go) for requirementSiteRegex, Generate test for this regexpt ✅ ([private chat](https://chatgpt.com/c/67aa31b3-85c8-800d-8237-686acd9ee06f))
- Copilot.Claude: Generate syntax errors constructions. Texts should be similaer to requirements ("shall") ❌
- ChatGPT o3-mini-high: Generate syntax errors constructors (e.g. NewErrPkgIdent). Error text should be similar to req text ("shall") ✅
- NI: errors.go. 1h.
- Copilot.Claude: Generate ParseCoverageFootnote ✅ (but do not work)
- Copilot.Claude: Generate TestMdParser_ParseCoverageFootnote ❌
- ChatGPT o3-mini-high: see Prompt_CoverageFootnoteRegex ✅
- regexps :( 1h
- Copilot.Claude: Rewrite the test using testify/assert👍
  - Refactor TestMdParser_ParseMarkdownFile to use testify/assert for improved readability and consistency
- Copilot.o3-mini
  - Implement "Test coverage footnote" block, see below👍
  - Implement ParseSourceFile using same approach as for ParseMarkdownFile👍
  - Generate TestMdParser_ParseSourceFile using same approach as for TestMdParser_ParseMarkdownFile: ✅ but no test data
  - Generate testdata/srccoverparser-1.go: file created but...❌
- Generate testdata/srccoverparser-1.go

**Next**. Some tests work, so continue

- If regexpt can match emojis: yes, TestRegexpEmojis
- Copilot.Claude
  - Align RequirementSiteRegex() and ParseRequirements() with ebnf notation. 👍 with minor flaws
  - Implement TestParseRequirements() using examples from TestRequirementSiteRegex().👍
  - Generate TestMdParser_ParseMarkdownFile_Errors() that parses #file:mdparser-errs.md and check all errors om this file.👍
  - mdparser.go shall identify pkgident, reqident and covstatus errors.👍🏆
- Implement FoldersScanner
  - Copilot.Claude: ❌ Much better prompt needed
  - ChatGPT o1: : ❌ Much better prompt needed
  - ChatGPT o3-mini-high: : ❌ Much better prompt needed
  - Claude 3.5: Process folders breadth-first, tests 👍🏆
    - Generate TestFoldersScanner_ALotOfErrors👍🏆

## scanner.go

- Context for ParseMarkdownFile to provide `ReqmdfilesMap`
- GitHub Copilot: 4o:Test_ParseCoverageFootnote: pass MarkdownContext with urls and check all results: works but had to be fixed
- Implement Scan function
  - Copilot.Claude:Server error: 500 Internal Server Error
  - Copilot.o1: ❌ (very bad)
  - Copilot.o3-mini: ❌
  - ChatGPT o3-minin-high: starting point
  - Claude 3.5: 👍🏆
- Copilot.Claude
  - Create a new function processSourceFile that takes all necessary parameters. Replace the anonymous function with a call to processSourceFile.👍

## Tasks

### tracer.go

- Oops. Refactor func Scan(paths []string) => func Scan(reqPath string, srcPaths []string)👍
  - Let me help you refactor the Scan function to accept separate requirement and source paths
  - Change the signature of Scan function and simplify its implementation since paths are now properly separated.
- Copilot.Claude  
  - Add parameters to NewTracer() to feed all interfaces and implement (t *tracer) Trace()❌
- Fix design
- Cleanup tracer.go
- Copilot.Claude: #file:tracer.go : construct tracer struct that implements ITracer. ITracer shall be created by NewTracer(all necessary params) ITracer methods.❌
- Copilot.o3-mini:: #file:tracer.go : construct tracer struct that implements ITracer. ITracer shall be created by NewTracer(all necessary params) ITracer methods.❌
- Change design
  - From: Depend on **abstractions** (`IScanner`, `IAnalyzer`, `IApplier`), not on concrete implementations.
  - To: Use injected interfaced (ref. interfaces.go) IScanner, IAnalyzer, IApplier to scan, analyze, and apply changes.
- Copilot.Claude: #file:tracer.go : construct tracer struct that implements ITracer. ITracer shall be created by NewTracer(all necessary params) ITracer methods.👍

#### tracer.go design

- ChatGPT [o3-mini-high](https://chatgpt.com/g/g-p-675ed7110c588191ae3694ae50086ad0-reqmd/c/67ae5cd3-ece4-800d-b811-bb40e0c32795): Prompt1.❌
- ChatGPT [o1](https://chatgpt.com/c/67ae5d38-8a38-800d-94b7-4419f51c6d28): Prompt1.❌
- Claude: [Prompt1](https://claude.ai/chat/0448fe13-c6ec-4271-860b-290e13ae7c36).👍 Way better, but prompt should be amended.

Prompt1

```text
Suggest a design for the implementation of the IAnalyzer that meets the fillowing requirements. Modify Action type if needed. Follow SOLID principles.

    - Determine which coverage footnotes need to be updated or created.  
    - Identify which bare requirement names need coverage annotations appended.  
    - Compare file hashes in reqmdfiles.json to actual git hash-object results to see if coverage references are stale.  
    - Construct the list of Action items describing needed transformations.
```

### oops, URL calculation is missing (+2h)

- There should be a way to obtain FileURL for a given FileStructure
- Me: Edit design.md, FileURL construction

- Copilot.Claude: Is it clear how to implement ## FileURL construction from #file:design.md ? Ask questions if needed.🏆
  - Answer:
    - Remote URL: How do we determine the repository's remote URL (e.g., "https://github.com/voedger/voedger")? This seems necessary to construct the RepoRootFolderURL. Should this be obtained through the Git interface?
    - Branch Name: The examples show URLs with "/blob/main/". Should we: Hardcode "main" as the branch name? Get it from the current branch? Make it configurable?
    - Platform Detection: The examples show different URL formats: GitHub: ".../blob/..." GitLab: ".../-/blob/..." How should we detect which platform format to use?
    - Branch Name in URLs The design has conflicting information:🏆
      - Git provider-specific path elements:
        - GitHub: `blob/main`
        - GitLab: `-/blob/main`
      - But also mentions: Actual current branch name

- Me: Prepare design of the ## File URL construction
- Claude: Rewrite this part professionally. Keep ALL names intact.👍

- loop ~10 times
  - Copilot.Claude: Is it clear how to implement ## FileURL construction from #file:design.md ? Ask quiestions if needed
  - Me: Modify design
- Copilot.Claude: Based on your answers, I'll propose the implementation. We need to modify several files🏆

### cleanup

- Copilot.Claude: Identify LLMs notes that should be removed👍
  - The following line appears twice in the document and should be removed: Let me help you rewrite this technical documentation with a more professional structure and better formatting.
  - They are clearly meta-comments from an AI assistant and not part of the actual technical documentation. The content before and after these lines is legitimate design documentation and should remain in place.
- Copilot.Claude: Review the design:👍

### Implement main(): instantiate all necessary components and run Tracer

Yes, so far IApplier is not implemented, so we should prepare a dummy implementation, but the rest can be tested already.

- Copilot.Claude: Prepare a dummy implementation of IApplier in applier.go.
  - Failed first, had to add more to context☝️
- Copilot.Claude: Rewrite like git documentation, use unordered lists for identation, do not use spaces.🏆
  - Context: requirements.md, `### Tracing` section
- Me: Oops, analyzer is missed.
- Copilot.Claude: Implement dummy analyzer in analyzer.go. Ref. #file:interfaces.go , #file:models.go #file:scanner.go #file:applier.go #file:tracer.go👍
- Copilot.Claude
- Copilot.Claude: Is it clear how to implement ### Tracing secction from the #file:requirements.md ? Ref. #interfaces.go  ,#models.go  #scanner.go  #applier.go  #tracer.go👍
- Me: Apply changes to main.go
- Copilot.Claude: rewrite main.go logic using cobra👍

Prompt

```text
Implement dummy analyzer in analyzer.go. Ref. #interfaces.go , #models.go #scanner.go #applier.go #tracer.go
```

### Run tests and make sure that the scanner works properly

- Me: it is necessary to process only specified file extensions in sources
- Copilot.Claude: I want to pass a list of source file extensions as argument s for reqmd trace, what you recommend?👍
- Me: Update the requirements.md
- Copilot.Claude: Fix design.md and scanner.go according to requirements.md (source file extensions)❌
  - Cause: main.go is missed in context
- Copilot.Claude: Fix design.md, main.go and scanner.go according to changes in requirements.md (source file extensions)👍

### Large source files are avoided

Input files that are larger than 128KB are not processed.

- Copilot.Claude: Apply requirements to scanner:
  - Files that are larger than 128K are skipped.
  - Only source files that are tracked by git (hash can be obtained) are processed.

- Copilot.Claude: scanner should Verbose number of encountered files and their total size, and number of skipped files and theit total size.👍

### Only one RequirementSite is allowed per line

- Copilot.Claude: Only one RequirementSite is allowed per line.🏆
  - Context: requirements.md, errors.go, mdparser.go
  - See the response in the Intermediate results section
- Copilot.Claude: Cover NewErrMultiSites by new TestParseRequirements_errors🏆

### ActionAnnotate

- Copilot.Claude: Generate Action.String() method: OK
- Copilot.Claude: Is it clear how to process ActionAnnotate in applier.go. It is necessary to annotate RequirementSite and add a coverage footnote. Use RequirementSiteRegex to identify RequirementSite. Ask questions, if needed.🏆
  - Context: #file:applier.go #file:requirements.md #file:ebnf.md #file:mdparser.go
- Copilot.Claude: Is it clear how to implement `## ActionAnnotate processing` section  from the design.md? Ask questions, if needed.
  - Context: #file:applier.go #file:requirements.md #file:ebnf.md #file:mdparser.go
- Me: repeat the question multiple times. Response is incredibly cool🏆. An example:
  - When Action.Type is ActionAdd for footnotes:
    - Does Action.Data contain the complete footnote text?
    - Or do we need to construct it from some components?
    - Should we maintain any specific order of footnotes?

--------------------

## Intermediate results

### Prompt_CoverageFootnoteRegex

```text
Why 

CoverageFootnoteRegex = regexp.MustCompile(^\s*\[^~([^~]+)~\]:\s* + "")

does not match

line := "[^~REQ002~]: [~com.example.basic~impl][folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)..."
```

### Test coverage footnote

```go
  // Test coverage footnote
  // [^~REQ002~]: `[~com.example.basic/REQ002~impl]` [folder1/filename1:line1:impl](https://example.com/pkg1/filename1), [folder2/filename2:line2:test](https://example.com/pkg2/filename2)
	{
		assert.Len(t, basicFile.CoverageFootnotes, 1, "should have 1 coverage footnote")
		if len(basicFile.CoverageFootnotes) > 0 {
			footnote := basicFile.CoverageFootnotes[0]
			assert.Equal(t, "REQ002", footnote.RequirementID, "incorrect requirement ID in footnote")
			assert.Equal(t, "com.example.basic", footnote.PackageID, "incorrect package ID in footnote")

			require.Len(t, footnote.Coverers, 2, "should have 2 coverage references")
			assert.Equal(t, "folder1/filename1:line1:impl", footnote.Coverers[0].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg1/filename1", footnote.Coverers[0].CoverageURL)
			assert.Equal(t, "folder2/filename2:line2:test", footnote.Coverers[1].CoverageLabel)
			assert.Equal(t, "https://example.com/pkg2/filename2", footnote.Coverers[1].CoverageURL)
		}
	}
}
```

### Response: Only one RequirementSite is allowed per line

I'll help implement the "one RequirementSite per line" requirement. Here's the step-by-step solution:

1. Add a new error type for multiple requirements per line
2. Modify the ParseRequirements function to check for multiple matches
3. Add the corresponding error constructor in errors.go

#### [mdparser.go](file:///c%3A/workspaces/work/reqmd/internal/mdparser.go)

Modify ParseRequirements to check for multiple matches per line.

```go
// ...existing code...

func ParseRequirements(filePath string, line string, lineNum int, errors *[]ProcessingError) []RequirementSite {
	var requirements []RequirementSite

	matches := RequirementSiteRegex.FindAllStringSubmatch(line, -1)
	if len(matches) > 1 {
		*errors = append(*errors, NewErrMultipleRequirements(filePath, lineNum))
		return nil
	}
	
	for _, match := range matches {
		reqName := match[1]
		if !identifierRegex.MatchString(reqName) {
			*errors = append(*errors, NewErrReqIdent(filePath, lineNum))
		}
		// ...existing code...
	}

	return requirements
}

// ...existing code...
```

#### [errors.go](file:///c%3A/workspaces/work/reqmd/internal/errors.go)

Add the new error constructor for multiple requirements per line.

```go


Made changes.

// ...existing code...

func NewErrMultipleRequirements(filePath string, line int) ProcessingError {
	return ProcessingError{
		Code:     "multireq",
		FilePath: filePath,
		Line:     line,
		Message:  "Only one RequirementSite is allowed per line",
	}
}

// ...existing code...
```

The changes ensure that:

1. If multiple requirements are found on a single line, an error is reported
2. No requirements from that line are processed (returning nil)
3. The error message clearly indicates the violation of the one-requirement-per-line rule

Made changes.

### Response: Cover NewErrMultiSites by new TestParseRequirements_errors

I'll help add a test function to cover the NewErrMultiSites error case. Here's the step-by-step solution:

1. Create a new test function TestParseRequirements_errors
2. Include test cases for multiple requirement sites on a single line

```go
package internal

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseRequirements_errors(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantErr  ProcessingError
	}{
		{
			name: "multiple requirements on single line",
			line: "`~REQ001~` `~REQ002~`",
			wantErr: ProcessingError{
				Code:     "multisites",
				FilePath: "test.md",
				Line:     1,
				Message:  "Only one RequirementSite is allowed per line: `~REQ001~`,  `~REQ002~`",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errors []ProcessingError
			requirements := ParseRequirements("test.md", tt.line, 1, &errors)

			// Check that no requirements were returned
			assert.Nil(t, requirements, "should return nil requirements when multiple sites found")

			// Verify the error
			if assert.Len(t, errors, 1, "should have exactly one error") {
				assert.Equal(t, tt.wantErr.Code, errors[0].Code)
				assert.Equal(t, tt.wantErr.Message, errors[0].Message)
				assert.Equal(t, tt.wantErr.FilePath, errors[0].FilePath)
				assert.Equal(t, tt.wantErr.Line, errors[0].Line)
			}
		})
	}
}
```

The test:

1. Defines a test case with multiple requirement sites on a single line
2. Verifies that ParseRequirements returns nil for requirements
3. Checks that the correct error is generated with the expected properties
4. Uses the testify/assert package for cleaner assertions

You can add this test function to your existing mdparser_test.go file.

Made changes.

### Response: Is it clear how to process ActionAnnotate in applier.go

I have some questions about implementing ActionAnnotate in applier.go:

1. For line modification:
   - When we find a RequirementSite using RequirementSiteRegex, should we:
     - Replace just the bare requirement (e.g., "`~REQ001~`") with the annotated version (e.g., "`~REQ001~`covered[^~REQ001~]✅"), or
     - Preserve any existing coverage status and just update it?

2. For coverage footnotes:
   - Where exactly should new coverage footnotes be added in the file:
     - At the very end of the file?
     - After the last existing footnote?
     - After the section containing the requirement?

3. File handling:
   - Should we read the entire file into memory and modify it, or
   - Process it line by line with a scanner/writer?
   - Do we need to preserve original line endings (CRLF vs LF)?

4. Error handling:
   - What should happen if:
     - The file is read-only?
     - Multiple RequirementSites are found in a line (though this should be caught earlier)?
     - The line number in the action doesn't match where we find the requirement?

5. Backup:
   - Should we create a backup of the original file before modifying it?
   - Or should this be handled at a higher level?

Once these points are clarified, I can propose a specific implementation for the applier.go file.

## Analysis

- Presumably: ChatGPT-4 produces better quality text content

### Cool prompts

- Copilot.Claude: Is it clear how to implement ### Tracing secction from the #file:requirements.md ? Ref. #interfaces.go  ,#models.go  #scanner.go  #applier.go  #tracer.go
- Copilot.Claude: Implement dummy analyzer in analyzer.go. Ref. #interfaces.go , #models.go #scanner.go #applier.go #tracer.
