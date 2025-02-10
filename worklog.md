# Worklog

This file contains a brief log of the project design and construction.

## LLMs

- Attempt 1. ChatGPT o3-mini-high. [private chat](https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed)
  - Prompt: Suggest architecture of the solution (Go). Do not generate all code yet. Provide list of files, key functions, structs and their resposibilities.
    - After the analysys of the output the Construction requirements section was added.
- Attempt 2. ChatGPT o3-mini-high. https://chatgpt.com/c/67a90782-3644-800d-a619-956119cc2b0c (not public)
  - Suggest the architecture of the solution using SOLID principles. Don't generate all the code yet. Provide a list of files, key functions, structures, and their responsibilities.
  - Generate internal/models.go
  - Generate internal/interfaces.go
    - Fix package names
  - Note: FileStructure was not defined
  - ChatGPT-4:
    - Prompt: Suggest the architecture of the solution using SOLID principles. Don't generate any code yet. Provide a list of files, key functions, structures, and their responsibilities.
    - Generate models.go and interfaces.go
  - GitHub Copilot: Suggest mdparser implementation ❌
  - Claude: Implement mdparser.go ✅
  - GitHub Copilot: Generate tests and testdata for mdparser.go
  - ChatGPT-3: Generate engaging README.md for this project ❌
  - ChatGPT-4: Generate engaging README.md for this project ✅
  - Manually: Improve mdparser_test.go
  - GitHub Copilot: Claude: Generate definition for requirementSiteRegex ❌
  - ChatGPT o3-mini-high: Generate definition (Go) for requirementSiteRegex, Generate test for this regexpt ✅ ([private chat](https://chatgpt.com/c/67aa31b3-85c8-800d-8237-686acd9ee06f))

Analysis

- Presumably: ChatGPT-4 produces better quality text content
