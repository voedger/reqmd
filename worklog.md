# Worklog

This file contains a brief log of the project design and construction.

## LLMs

**Attempt 1**. ChatGPT o3-mini-high. [private chat](https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed)

- Prompt: Suggest architecture of the solution (Go). Do not generate all code yet. Provide list of files, key functions, structs and their resposibilities.
  - After the analysys of the output the Construction requirements section was added.

**Attempt 2**. ChatGPT o3-mini-high. [private chat](https://chatgpt.com/c/67a90782-3644-800d-a619-956119cc2b0c)

- Suggest the architecture of the solution using SOLID principles. Don't generate all the code yet. Provide a list of files, key functions, structures, and their responsibilities.
- Generate internal/models.go
- Generate internal/interfaces.go
  - Fix package names
- Note: FileStructure was not defined
- ChatGPT o1:
  - Prompt: Suggest the architecture of the solution using SOLID principles. Don't generate any code yet. Provide a list of files, key functions, structures, and their responsibilities.
  - Generate models.go and interfaces.go
- GitHub Copilot: Suggest mdparser implementation ❌
- GitHub Copilot: Claude: Implement mdparser.go ✅
- GitHub Copilot: Claude: Generate tests and testdata for mdparser.go
- ChatGPT o1: Generate engaging README.md for this project ❌
- ChatGPT 4o: Generate engaging README.md for this project ✅
- NI: Improve mdparser_test.go
- GitHub Copilot: Claude: Generate definition for requirementSiteRegex ❌
- ChatGPT o3-mini-high: Generate definition (Go) for requirementSiteRegex, Generate test for this regexpt ✅ ([private chat](https://chatgpt.com/c/67aa31b3-85c8-800d-8237-686acd9ee06f))
- GitHub Copilot: Claude: Generate syntax errors constructions. Texts should be similaer to requirements ("shall") ❌
- ChatGPT o3-mini-high: Generate syntax errors constructors (e.g. NewErrPkgIdent). Error text should be similar to req text ("shall") ✅
- NI: errors.go. 1h.
- GitHub Copilot: Claude: Generate ParseCoverageFootnote ✅ (but do not work)
- GitHub Copilot: Claude: Generate TestMdParser_ParseCoverageFootnote ❌
- ChatGPT o3-mini-high: see Prompt_CoverageFootnoteRegex ✅
- regexps :( 1h
- GitHub Copilot: Claude: Rewrite the test using testify/assert👍
  - Refactor TestMdParser_ParseMarkdownFile to use testify/assert for improved readability and consistency
- GitHub Copilot: o3-mini
  - Implement "Test coverage footnote" block, see below👍
  - Implement ParseSourceFile using same approach as for ParseMarkdownFile👍
  - Generate TestMdParser_ParseSourceFile using same approach as for TestMdParser_ParseMarkdownFile: ✅ but no test data
  - Generate testdata/srccoverparser-1.go: file created but...❌
- Generate testdata/srccoverparser-1.go

**Next**. Some tests work, so continue

- If regexpt can match emojis: yes, TestRegexpEmojis
- GitHub Copilot: Claude
  - Align RequirementSiteRegex() and ParseRequirements() with ebnf notation. 👍 with minor flaws
  - Implement TestParseRequirements() using examples from TestRequirementSiteRegex().👍
  - Generate TestMdParser_ParseMarkdownFile_Errors() that parses #file:mdparser-errs.md and check all errors om this file.👍
  - mdparser.go shall identify pkgident, reqident and covstatus errors.👍🏆
- Implement FoldersScanner
  - GitHub Copilot: Claude: ❌ Much better prompt needed
  - ChatGPT o1: : ❌ Much better prompt needed
  - ChatGPT o3-mini-high: : ❌ Much better prompt needed
  - Claude 3.5: Process folders breadth-first, tests 👍🏆
    - Generate TestFoldersScanner_ALotOfErrors👍🏆
- scanner

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

## Analysis

- Presumably: ChatGPT-4 produces better quality text content
