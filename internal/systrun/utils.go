package systrun

import "regexp"

// Define the prefix for golden annotations
const goldenAnnotationPrefix = ">"
const goldenAnnotationRegexpPrefix = `^\` + goldenAnnotationPrefix + `\s*`

func ga(line string) string {
	return goldenAnnotationPrefix + " " + line
}

func gare(line string) *regexp.Regexp {
	return regexp.MustCompile(goldenAnnotationRegexpPrefix + line)
}
