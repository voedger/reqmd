// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"strings"
	"time"
)

var IsVerbose bool

func Verbose(msg string, kv ...any) {
	if !IsVerbose {
		return
	}

	Log(msg, kv...)
}

func Log(msg string, kv ...any) {
	pairs := make([]string, 0, len(kv)/2)
	for i := 0; i < len(kv)-1; i += 2 {
		key := quoteIfNeeded(fmt.Sprint(kv[i]))
		val := quoteIfNeeded(fmt.Sprint(kv[i+1]))
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, val))
	}

	if len(kv)%2 != 0 {
		pairs = append(pairs, fmt.Sprint(kv[len(kv)-1]))
	}

	timestamp := time.Now().Format("15:04:05.000")
	if len(pairs) > 0 {
		fmt.Printf("%s %s: %s\n", timestamp, msg, strings.Join(pairs, " "))
	} else {
		fmt.Printf("%s %s\n", timestamp, msg)
	}
}

// quoteIfNeeded returns the string quoted if it contains spaces, tabs, or double quotes, otherwise returns as is
func quoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t\"") {
		escaped := strings.ReplaceAll(s, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return s
}
