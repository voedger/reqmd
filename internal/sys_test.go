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
	runSysTest(t, "noreqs")
}

func Test_systest_errors_syn(t *testing.T) {
	runSysTest(t, "errors_syn")
}

func Test_systest_errors_sem(t *testing.T) {
	runSysTest(t, "errors_sem")
}

func Test_systest_justreqs(t *testing.T) {
	runSysTest(t, "justreqs")
}

// Reqs and srcs in different folders
func Test_systest_req_src(t *testing.T) {
	runSysTest(t, "req_src")
}

// Requirements and sources in the same folder
func Test_systest_reqsrc(t *testing.T) {
	runSysTest(t, "reqsrc")
}

func runSysTest(t *testing.T, testID string) {
	systrun.RunSysTest(t, sysTestsDir, testID, ExecRootCmd, Version)
}
