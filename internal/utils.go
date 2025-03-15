// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"log"
	"strings"
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
		pairs = append(pairs, fmt.Sprintf("\n\t%v: %v", kv[i], kv[i+1]))
	}

	if len(kv)%2 != 0 {
		pairs = append(pairs, fmt.Sprintf("\n\t%v: <missing>", kv[len(kv)-1]))
	}

	if len(pairs) > 0 {
		log.Printf("%s %s", msg, strings.Join(pairs, ", "))
	} else {
		log.Println(msg)
	}
}
