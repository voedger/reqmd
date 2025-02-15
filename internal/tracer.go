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
	files, procErrs, err := t.scanner.Scan(t.reqPath, t.srcPaths)
	if err != nil {
		return err
	}
	if len(procErrs) > 0 {
		return &ProcessingErrors{Errors: procErrs}
	}

	// Analyzing phase
	actions, procErrs := t.analyzer.Analyze(files)
	if len(procErrs) > 0 {
		return &ProcessingErrors{Errors: procErrs}
	}

	// Applying phase
	if err := t.applier.Apply(actions); err != nil {
		return err
	}

	return nil
}
