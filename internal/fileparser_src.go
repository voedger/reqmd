// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"regexp"
)

// Global regex for parsing source file coverage tags.
// A CoverageTag is expected in the form: [~PackageId/RequirementName~CoverageType]
var coverageTagRegex = regexp.MustCompile(`(?:[^` + "`" + `])\[\~([^/]+)/([^~]+)\~([^\]]+)\]`)
