package internal

// NewErrPkgIdent creates a SyntaxError indicating that a PackageName is not a valid identifier.
// According to the specification, "PackageName shall be an identifier".
func NewErrPkgIdent(filePath string, line int) SyntaxError {
	return SyntaxError{
		Code:     "pkgident",
		FilePath: filePath,
		Line:     line,
		Message:  "PackageName shall be an identifier",
	}
}

// NewErrReqIdent creates a SyntaxError indicating that a RequirementName is not a valid identifier.
// According to the specification, "RequirementName shall be an identifier".
func NewErrReqIdent(filePath string, line int) SyntaxError {
	return SyntaxError{
		Code:     "reqident",
		FilePath: filePath,
		Line:     line,
		Message:  "RequirementName shall be an identifier",
	}
}
