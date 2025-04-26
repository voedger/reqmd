// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import "fmt"

// ********** Semantic errors

func NewErrDuplicateRequirementId(filePath1 string, line1 int, filePath2 string, line2 int, reqId RequirementId) ProcessingError {
	return ProcessingError{
		Code:     "dupreqid",
		FilePath: filePath1,
		Line:     line1,
		Message: fmt.Sprintf("duplicate RequirementId detected:\n\t%s\n\t%s:%d",
			reqId, filePath2, line2),
	}
}

func NewErrMissingPackageIdWithReqs(filePath string, lineOfTheFirstReqSite int) ProcessingError {
	return ProcessingError{
		Code:     "nopkgidreqs",
		FilePath: filePath,
		Line:     lineOfTheFirstReqSite,
		Message:  "markdown file with RequirementSites shall define reqmd.package",
	}
}

func NewErrPkgMismatch(filePath string, line int, pkgId1, pkgId2 string) ProcessingError {
	return ProcessingError{
		Code:     "pkgmismatch",
		FilePath: filePath,
		Line:     line,
		Message:  fmt.Sprintf("CoverageFootnote package (%s) is not consistent with PackageId in the header (%s)", pkgId1, pkgId2),
	}
}
