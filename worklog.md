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

## tracer.go

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

### tracer.go design

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

## oops, URL calculation is missing (+2h)

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

## cleanup

- Copilot.Claude: Identify LLMs notes that should be removed👍
  - The following line appears twice in the document and should be removed: Let me help you rewrite this technical documentation with a more professional structure and better formatting.
  - They are clearly meta-comments from an AI assistant and not part of the actual technical documentation. The content before and after these lines is legitimate design documentation and should remain in place.
- Copilot.Claude: Review the design:👍

## Implement main(): instantiate all necessary components and run Tracer

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

## Run tests and make sure that the scanner works properly

- Me: it is necessary to process only specified file extensions in sources
- Copilot.Claude: I want to pass a list of source file extensions as argument s for reqmd trace, what you recommend?👍
- Me: Update the requirements.md
- Copilot.Claude: Fix design.md and scanner.go according to requirements.md (source file extensions)❌
  - Cause: main.go is missed in context
- Copilot.Claude: Fix design.md, main.go and scanner.go according to changes in requirements.md (source file extensions)👍

## Format error output

- Me: Similar to Go compiler output

## Large source files are avoided

Input files that are larger than 128KB are not processed.

- Copilot.Claude: Apply requirements to scanner:
  - Files that are larger than 128K are skipped.
  - Only source files that are tracked by git (hash can be obtained) are processed.

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

## Analysis

- Presumably: ChatGPT-4 produces better quality text content

### Cool prompts

- Copilot.Claude: Is it clear how to implement ### Tracing secction from the #file:requirements.md ? Ref. #interfaces.go  ,#models.go  #scanner.go  #applier.go  #tracer.go
- Copilot.Claude: Implement dummy analyzer in analyzer.go. Ref. #interfaces.go , #models.go #scanner.go #applier.go #tracer.
