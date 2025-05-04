# Phases

Files are processed in three phases:

- Scan
  - Parse all InputFiles and generate FileStructures and the list of SyntaxErrors
  - InputFiles shall be processed per-subfolder by the goroutines pool
- Analyze
  - Preconditions: there are no SyntaxErrors
  - Parse all FileStructures and generate list of SemanticErrors and list of Actions
- Apply
  - Preconditions: there are no SemanticErrors
  - Apply all Actions to the InputFiles
