// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"path/filepath"
	"testing"

	"github.com/voedger/reqmd/internal/systrun"
)

var sysTestsDir = filepath.Join("testdata", "systest")

func Test_systest_noreqs(t *testing.T) {
	systrun.RunSysTest(t, sysTestsDir, "noreqs", ExecRootCmd, []string{"trace"}, "0.0.1")
}

func Test_systest_errors(t *testing.T) {
	systrun.RunSysTest(t, sysTestsDir, "errors", ExecRootCmd, []string{"trace"}, "0.0.1")
}

func Test_systest_justreqs(t *testing.T) {
	systrun.RunSysTest(t, sysTestsDir, "justreqs", ExecRootCmd, []string{"trace"}, "0.0.1")
}

// Reqs and srcs in different folders
func Test_systest_req_src(t *testing.T) {
	systrun.RunSysTest(t, sysTestsDir, "req_src", ExecRootCmd, []string{"trace"}, "0.0.1")
}

// Requirements and sources in the same folder
func Test_systest_reqsrc(t *testing.T) {
	systrun.RunSysTest(t, sysTestsDir, "reqsrc", ExecRootCmd, []string{"trace"}, "0.0.1")
}
