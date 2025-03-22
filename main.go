// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "embed"
	"os"

	"github.com/voedger/reqmd/internal"
)

//go:embed version
var version string

func main() {
	if err := internal.ExecRootCmd(os.Args, version); err != nil {
		os.Exit(1)
	}
}
