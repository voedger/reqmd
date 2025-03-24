// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systest

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal"
)

var runSysTestsDir = filepath.Join("testdata", "runsystest")

func Test_err_NotOccurring(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_undetected", internal.ExecRootCmd, []string{"trace"}, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	require.Contains(t, mockT.failMsg, "Expected error not found in stderr")
	require.Contains(t, mockT.failMsg, "this error is expected but not occurring")
}

func Test_err_Unexpected(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_unexpected", internal.ExecRootCmd, []string{"trace"}, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	require.Contains(t, mockT.failMsg, "Unexpected error")
	require.Contains(t, mockT.failMsg, "PackageID shall be an identifier: 11com.example.basic")
}

// MockT implements a subset of testing.T for controlled failure testing
type MockT struct {
	t       *testing.T
	failed  bool
	failMsg string
}

func (m *MockT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = fmt.Sprintf(format, args...)
}

func (m *MockT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = fmt.Sprintf(format, args...)
	panic("Fatalf called: " + m.failMsg)
}

func (m *MockT) FailNow() {
	panic("FailNow called")
}

func (m *MockT) TempDir() string {
	return m.t.TempDir()
}
