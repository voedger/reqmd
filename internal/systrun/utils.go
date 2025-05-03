// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systrun

import "regexp"

// Define the prefix for golden annotations
const goldenAnnotationPrefix = "@"
const goldenAnnotationRegexpPrefix = `^\` + goldenAnnotationPrefix + `\s*`

func ga(line string) string {
	return goldenAnnotationPrefix + " " + line
}

func gare(line string) *regexp.Regexp {
	return regexp.MustCompile(goldenAnnotationRegexpPrefix + line)
}
