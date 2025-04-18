// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import "fmt"

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
func NewErrURLSyntax(filePath string, line int, URL string, err error) ProcessingError {
	return ProcessingError{
		Code:     "urlsyntax",
		FilePath: filePath,
		Line:     line,
		Message:  "the URL provided is invalid: " + err.Error() + ": " + URL,
	}
}

// Only one RequirementSite is allowed per line
func NewErrMultiSites(filePath string, line int, site1, site2 string) ProcessingError {
	return ProcessingError{
		Code:     "multisites",
		FilePath: filePath,
		Line:     line,
		Message:  "only one RequirementSite is allowed per line: " + site1 + ",  " + site2,
	}
}

// Unmatched code block fence detected
func NewErrUnmatchedFence(filePath string, openFenceLine int) ProcessingError {
	return ProcessingError{
		Code:     "unmatchedfence",
		FilePath: filePath,
		Line:     openFenceLine,
		Message:  fmt.Sprintf("opening code block fence at line %d has no matching closing fence", openFenceLine),
	}
}

// ********** Semantic errors

func NewErrDuplicateRequirementId(filePath1 string, line1 int, filePath2 string, line2 int, reqID RequirementId) ProcessingError {
	return ProcessingError{
		Code:     "dupreqid",
		FilePath: filePath1,
		Line:     line1,
		Message: fmt.Sprintf("duplicate RequirementId detected:\n\t%s\n\t%s:%d",
			reqID, filePath2, line2),
	}
}

func NewErrMissingPackageIDWithReqs(filePath string, lineOfTheFirstReqSite int) ProcessingError {
	return ProcessingError{
		Code:     "nopkgidreqs",
		FilePath: filePath,
		Line:     lineOfTheFirstReqSite,
		Message:  "markdown file with RequirementSites shall define reqmd.package",
	}
}
