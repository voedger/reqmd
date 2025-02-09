# LLMs usage

This file contains the log of the usage of LLMs for the project.

## LLMs

- Attempt 1. ChatGPT o3-mini-high. https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed
  - Prompt: Suggest architecture of the solution (Go). Do not generate all code yet. Provide list of files, key functions, structs and their resposibilities.
- Attempt 2. ChatGPT o3-mini-high. https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed
  - Suggest the architecture of the solution using SOLID principles. Don't generate all the code yet. Provide a list of files, key functions, structures, and their responsibilities.
  - Generate internal/models.go
  - Generate internal/interfaces.go
    - Fix package names
  - Oops, FileStructure was not defined
  - o1
  - Suggest the architecture of the solution using SOLID principles. Don't generate any code yet. Provide a list of files, key functions, structures, and their responsibilities.
  - Generate models.go and interfaces.go
  - Copilot: o1: Suggest mdparser implementation: :(
  - Copilot: claude: Implement mdparser.go: :)
  - Copilot: generate tests and testdata for mdparser.go
  - ChatGPT: o1: Generate cool README.md for this project: :(
  - ChatGPT: 4o: Generate cool README.md for this project: :)

Analysis

- "Suggest the design of the solution", not "architecture".
- ChatGPT 4o is better for generating texts?
