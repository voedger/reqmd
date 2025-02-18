package internal

/*

An exerpt from design.md

### **tracer.go**

- **Purpose**: Implements `ITracer`. This is the **facade** that orchestrates the scanning, analyzing, and applying phases.
- **Key functions**:
  - `Trace()
- **Responsibilities**:
  - High-level workflow control.
  - Enforce the steps: if syntax errors exist, abort; if semantic errors exist, abort; otherwise apply actions.
  - Use injected interfaced (ref. interfaces.go) IScanner, IAnalyzer, IApplier to scan, analyze, and apply changes.

*/

type tracer struct {
	scanner  IScanner
	analyzer IAnalyzer
	applier  IApplier
	reqPath  string
	srcPaths []string
}

func NewTracer(scanner IScanner, analyzer IAnalyzer, applier IApplier, reqPath string, srcPaths []string) ITracer {
	return &tracer{
		scanner:  scanner,
		analyzer: analyzer,
		applier:  applier,
		reqPath:  reqPath,
		srcPaths: srcPaths,
	}
}

func (t *tracer) Trace() error {
	// Scanning phase
	scanResult, err := t.scanner.Scan(t.reqPath, t.srcPaths)
	if err != nil {
		return err
	}
	if len(scanResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: scanResult.ProcessingErrors}
	}

	// Analyzing phase
	analyzeResult, err := t.analyzer.Analyze(scanResult.Files)
	if err != nil {
		return err
	}
	if len(analyzeResult.ProcessingErrors) > 0 {
		return &ProcessingErrors{Errors: analyzeResult.ProcessingErrors}
	}

	// Applying phase
	if err := t.applier.Apply(analyzeResult.Actions); err != nil {
		return err
	}

	return nil
}
