package internal

// ********** Syntax errors

func NewErrPkgIdent(filePath string, line int) SyntaxError {
	return SyntaxError{
		Code:     "pkgident",
		FilePath: filePath,
		Line:     line,
		Message:  "PackageName shall be an identifier",
	}
}

func NewErrReqIdent(filePath string, line int) SyntaxError {
	return SyntaxError{
		Code:     "reqident",
		FilePath: filePath,
		Line:     line,
		Message:  "RequirementName shall be an identifier",
	}
}

func NewErrRequirementSiteIDEqual(filePath string, line int, RequirementSiteID1, RequirementSiteID2 string) SyntaxError {
	return SyntaxError{
		Code:     "reqsiteid",
		FilePath: filePath,
		Line:     line,
		Message:  "RequirementSiteID from RequirementSiteLabel and CoverageFootnoteReference shall be equal: " + RequirementSiteID1 + " != " + RequirementSiteID2,
	}
}

// CoverageStatusWord shall be "covered" or "uncvrd"
func NewErrCoverageStatusWord(filePath string, line int, CoverageStatusWord string) SyntaxError {
	return SyntaxError{
		Code:     "covstatus",
		FilePath: filePath,
		Line:     line,
		Message:  "CoverageStatusWord shall be 'covered' or 'uncvrd':" + CoverageStatusWord,
	}
}

// URL shall adhere to a valid syntax
func NewErrURLSyntax(filePath string, line int, URL string) SyntaxError {
	return SyntaxError{
		Code:     "urlsyntax",
		FilePath: filePath,
		Line:     line,
		Message:  "URL shall adhere to a valid syntax: " + URL,
	}
}

// ********** Semantic errors
