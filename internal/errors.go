package internal

// ********** Syntax errors

func NewErrPkgIdent(filePath string, line int, pkgID string) ProcessingError {
	return ProcessingError{
		Code:     "pkgident",
		FilePath: filePath,
		Line:     line,
		Message:  "PackageID shall be an identifier: " + pkgID,
	}
}

func NewErrReqIdent(filePath string, line int) ProcessingError {
	return ProcessingError{
		Code:     "reqident",
		FilePath: filePath,
		Line:     line,
		Message:  "RequirementName shall be an identifier",
	}
}

func NewErrRequirementSiteIDEqual(filePath string, line int, RequirementSiteID1, RequirementSiteID2 string) ProcessingError {
	return ProcessingError{
		Code:     "reqsiteid",
		FilePath: filePath,
		Line:     line,
		Message:  "RequirementSiteID from RequirementSiteLabel and CoverageFootnoteReference shall be equal: " + RequirementSiteID1 + " != " + RequirementSiteID2,
	}
}

// CoverageStatusWord shall be "covered" or "uncvrd"
func NewErrCoverageStatusWord(filePath string, line int, CoverageStatusWord string) ProcessingError {
	return ProcessingError{
		Code:     "covstatus",
		FilePath: filePath,
		Line:     line,
		Message:  "CoverageStatusWord shall be 'covered' or 'uncvrd': " + CoverageStatusWord,
	}
}

// URL shall adhere to a valid syntax
func NewErrURLSyntax(filePath string, line int, URL string) ProcessingError {
	return ProcessingError{
		Code:     "urlsyntax",
		FilePath: filePath,
		Line:     line,
		Message:  "URL shall adhere to a valid syntax: " + URL,
	}
}

func NewErrMultiSites(filePath string, line int, site1, site2 string) ProcessingError {
	return ProcessingError{
		Code:     "multisites",
		FilePath: filePath,
		Line:     line,
		Message:  "Only one RequirementSite is allowed per line: " + site1 + ",  " + site2,
	}
}

// ********** Semantic errors
