# Design

Design of the reqmd tool.

## LLMs

- Attempt 1. ChatGPT o3-mini-high. https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed
  - Prompt: Suggest architecture of the solution (Go). Do not generate all code yet. Provide list of files, key functions, structs and their resposibilities.
- Attempt 2. ChatGPT o3-mini-high. https://chatgpt.com/c/67a7f223-fcc0-800d-a486-427a3f47c3ed
  - Prompt: Suggest the architecture of the solution using SOLID principles. Don't generate all the code yet. Provide a list of files, key functions, structures, and their responsibilities.

## Overview

The solution is designed as a modular command-line tool that traces requirements from Markdown files to their corresponding coverage in source files. It follows SOLID principles to ensure that each component has a single responsibility, is easily testable, and can be extended without modifying existing code. Here’s a high-level overview of the design:

- **Modular Architecture:**  
  The system is divided into several components, each responsible for a specific phase in the processing pipeline:
  - **Scanning:** Traverses directories, locates Markdown and source files, and parses them into structured models.
  - **Analysis:** Validates the parsed data for semantic correctness (e.g., uniqueness of RequirementIDs) and produces a list of actions that need to be performed.
  - **Application:** Executes the generated actions by updating Markdown files and the `reqmdfiles.json` file.
  - **File Hashing:** Calculates file hashes using the `git hash-object` command to determine if files have changed.

- **Central Orchestration with the Tracer:**  
  A dedicated tracer component acts as the central coordinator that orchestrates the complete workflow (Scan → Analyze → Apply). It implements the `ITracer` interface and is responsible for injecting dependencies, managing execution flow, and handling errors.

- **Interface-Driven Design:**  
  Key functionalities are abstracted into interfaces (e.g., `ITracer`, `IScanner`, `IAnalyzer`, `IApplicator`, `IFileHasher`) which:
  - **Ensure Dependency Inversion:** High-level modules depend on abstractions rather than concrete implementations.
  - **Promote Flexibility:** Concrete implementations can be easily swapped, extended, or mocked during testing.
  - **Encourage Interface Segregation:** Each interface is fine-grained, exposing only the methods required by its consumers.

- **Separation of Data Models:**  
  All shared data structures (such as Markdown files, requirements, coverage footnotes, and actions) are defined in a central `models.go` file. This provides a single source of truth and ensures consistent usage across all modules.

- **File Organization:**  
  - **`main.go`:** Serves as the entry point of the application, parsing command-line arguments, initializing dependencies, and starting the tracing process.
  - **`internal/`:** Contains all the core components and logic. Files such as `interfaces.go`, `models.go`, `scanner.go`, `analyzer.go`, `applicator.go`, `hasher.go`, and `tracer.go` are organized to reflect distinct phases and responsibilities of the application.

- **Adherence to SOLID Principles:**
  - **Single Responsibility:** Each file and component handles one specific aspect of the tracing process.
  - **Open/Closed:** The system can be extended with new functionalities (like additional parsing rules or output formats) without modifying the existing code.
  - **Liskov Substitution:** Implementations can be substituted with no adverse effects as long as they adhere to the defined interfaces.
  - **Interface Segregation:** Clients only depend on the methods they need from each interface.
  - **Dependency Inversion:** High-level modules (like the tracer) rely on abstract interfaces, promoting loose coupling between components.

Overall, this design ensures that the tool is robust, maintainable, and flexible enough to accommodate future enhancements or modifications while maintaining clear, testable boundaries between its components.

## File Structure

```text
reqmd/
├── main.go
└── internal/
    ├── interfaces.go
    ├── models.go
    ├── tracer.go
    ├── scanner.go
    ├── analyzer.go
    ├── applicator.go
    └── hasher.go
```

---

## File and Component Details

### **main.go**

- **Responsibilities:**
  - Parse command-line arguments (e.g., the `trace` command with `<path-to-markdowns>` and optional `<path-to-cloned-repo>` paths).
  - Initialize dependencies and inject them into the tracer (i.e., create concrete implementations of the interfaces).
  - Start the tracing process and handle any errors.
- **Key Functions:**
  - `main()`: Application entry point.
  - `parseArgs()`: Parses CLI arguments and returns the necessary paths.

---

### **internal/interfaces.go**

- **Responsibilities:**
  - Define all the interfaces that abstract the core functionalities of the tool.
  - Enforce naming conventions (interface names start with `I`).
- **Key Interfaces:**
  - **ITracer**
    - **Method:** `Trace(markdownPaths []string, repoPaths []string) error`
    - **Responsibility:** Orchestrates the entire tracing process by coordinating scanning, analyzing, and applying changes.
  - **IScanner**
    - **Method:** `Scan(directory string) ([]*models.MarkdownFile, error)`
    - **Responsibility:** Recursively traverses directories to find markdown and source files, and parses them.
  - **IParser** *(optional, or part of the scanner)*
    - **Method:** `ParseMarkdown(filePath string) (*models.MarkdownFile, error)`
    - **Responsibility:** Converts a markdown file’s content into structured data.
  - **IAnalyzer**
    - **Method:** `Analyze(files []*models.MarkdownFile) (actions []*models.Action, semanticErrors []*models.SemanticError, error)`
    - **Responsibility:** Validates file structures (e.g., uniqueness of RequirementIDs) and generates actions (e.g., updating requirement sites and footnotes).
  - **IApplicator**
    - **Method:** `Apply(actions []*models.Action) error`
    - **Responsibility:** Applies generated actions to update markdown files and `reqmdfiles.json`.
  - **IFileHasher**
    - **Method:** `Hash(fileURL string) (string, error)`
    - **Responsibility:** Calculates file hashes using the `git hash-object` command.

---

### **internal/models.go**

- **Responsibilities:**
  - Define all the data structures and models used across the application.
  - Serve as the single source of truth for shared types.
- **Key Structures:**
  - **MarkdownFile**
    - Represents a markdown file, including its header (with PackageID) and body.
  - **FileStructure**
    - Represents the parsed content of an input file (markdown or source).
  - **Requirement / RequirementSite**
    - Represents requirements in the document (including BareRequirementName and RequirementSite with coverage annotations).
  - **CoverageFootnote & Coverer**
    - Models for the generated footnotes and coverers (including CoverageLabel and CoverageURL).
  - **Action**
    - Describes an update action (Type: Add, Update, Delete) along with metadata (file path, line number, new data).
  - **SyntaxError** and **SemanticError**
    - Represent issues encountered during parsing and validation.

---

### **internal/tracer.go**

- **Responsibilities:**
  - Implement the `ITracer` interface.
  - Act as the central coordinator, following the three processing phases:
    - **Scan:** Parse input files into structured models.
    - **Analyze:** Validate models (e.g., ensure uniqueness of RequirementIDs) and produce a list of actions.
    - **Apply:** Execute the actions by updating the source markdown files and the `reqmdfiles.json`.
- **Key Functions:**
  - `NewTracer(scanner IScanner, analyzer IAnalyzer, applicator IApplicator, fileHasher IFileHasher) *Tracer`
    - Constructor that injects the required interfaces.
  - `Trace(markdownPaths []string, repoPaths []string) error`
    - Executes the overall tracing process.

---

### **internal/scanner.go**

- **Responsibilities:**
  - Implement the `IScanner` interface.
  - Scan directories to discover markdown files and source files.
  - Delegate parsing of files to helper functions or an embedded parser.
- **Key Functions:**
  - `Scan(directory string) ([]*models.MarkdownFile, error)`
    - Traverses the given directory and collects markdown files.
  - `parseMarkdownFile(filePath string) (*models.MarkdownFile, error)`
    - Reads a markdown file and converts it into a `MarkdownFile` model.
  - *(Optional)* Functions to parse source files for CoverageTags.

---

### **internal/analyzer.go**

- **Responsibilities:**
  - Implement the `IAnalyzer` interface.
  - Analyze the parsed file structures to:
    - Detect any semantic errors (e.g., duplicate RequirementIDs).
    - Generate a list of actions needed to transform BareRequirementNames into RequirementSites and update CoverageFootnotes.
- **Key Functions:**
  - `Analyze(files []*models.MarkdownFile) (actions []*models.Action, semanticErrors []*models.SemanticError, error)`
    - Performs validation and generates update actions.
  - Helper functions to validate:
    - Uniqueness of RequirementIDs.
    - Correct formation of CoverageFootnotes.

---

### **internal/applicator.go**

- **Responsibilities:**
  - Implement the `IApplicator` interface.
  - Apply the actions generated during the analysis phase by modifying the markdown files and updating `reqmdfiles.json`.
- **Key Functions:**
  - `Apply(actions []*models.Action) error`
    - Iterates through and applies each action.
  - `updateMarkdownFile(filePath string, actions []*models.Action) error`
    - Updates a specific markdown file according to the action list.
  - `updateReqmdFilesJson(directory string, fileHashes map[string]string) error`
    - Creates or updates the `reqmdfiles.json` file with the latest file hashes.

---

### **internal/hasher.go**

- **Responsibilities:**
  - Implement the `IFileHasher` interface.
  - Calculate file hashes by interfacing with the `git hash-object` command.
- **Key Functions:**
  - `Hash(fileURL string) (string, error)`
    - Returns the hash for a file given its URL, used to determine if a file has been modified.

---

## SOLID Principles Recap

- **Single Responsibility:**  
  Each module (scanner, analyzer, applicator, etc.) has one clear task. For example, `scanner.go` only deals with file discovery and parsing, while `analyzer.go` focuses solely on validating and generating actions.

- **Open/Closed:**  
  The use of interfaces (e.g., `ITracer`, `IScanner`, `IAnalyzer`, etc.) allows the system to be extended (for instance, adding new types of analysis or different file parsers) without modifying existing code.

- **Liskov Substitution:**  
  Since high-level modules (like the Tracer) depend on abstractions (interfaces), any compliant implementation can be substituted in tests or for new features.

- **Interface Segregation:**  
  The interfaces are fine-grained. Consumers depend only on the methods they actually use (for instance, the tracer only requires the methods defined in `IScanner`, `IAnalyzer`, etc., rather than a large, monolithic interface).

- **Dependency Inversion:**  
  High-level modules (like the Tracer) depend on abstractions rather than concrete implementations. All dependencies are injected during construction (e.g., via `NewTracer`), promoting flexibility and easier testing.