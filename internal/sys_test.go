// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"path/filepath"
	"testing"

	"github.com/voedger/reqmd/internal/systest"
)

var sysTestsDir = filepath.Join("testdata", "systest")

func Test_systest_noreqs(t *testing.T) {
	systest.RunSysTest(t, sysTestsDir, "noreqs", ExecRootCmd, []string{"trace"}, "0.0.1")
}

func Test_systest_errors(t *testing.T) {
	systest.RunSysTest(t, sysTestsDir, "errors", ExecRootCmd, []string{"trace"}, "0.0.1")
}

// func Test_systest_justreqs(t *testing.T) {
// 	systest.RunSysTest(t, sysTestsDir, "justreqs", ExecRootCmd, []string{"trace"}, "0.0.1")
// }
