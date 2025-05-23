// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systrun

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal"
)

var runSysTestsDir = filepath.Join("testdata")

func Test_noerr(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "noerr", internal.ExecRootCmd, "0.0.1")
	assert.False(t, mockT.failed, mockT.String())
}

func Test_err_NotOccurring(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_notoccuring", internal.ExecRootCmd, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	mockT.assertMsgsContains("Expected error not found in stderr")
	mockT.assertMsgsContains("this error is expected but not occurring")
}

// Errors are not declared but occurr
func Test_err_Unexpected(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_unexpected", internal.ExecRootCmd, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	mockT.assertMsgsContains("Unexpected error")
	mockT.assertMsgsContains("PackageId shall be an identifier: 11com.example.basic")
}

// Errors are declared and occur but not matched
func Test_err_MatchedAndUnmatched(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_matchedunmatched", internal.ExecRootCmd, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	mockT.assertMsgsContains("Expected error not found in stderr")
	mockT.assertMsgsContains("Unexpected error")
	mockT.assertMsgsContains("PackageId shall be an identifier: 11com.example.basic")
}

func Test_err_LineCountMismatch(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_linecountmismatch", internal.ExecRootCmd, "0.0.1")
	require.True(t, mockT.failed, "expected test to fail")
	require.Contains(t, mockT.failMsg, "Line count mismatch in req.md: expected 1 lines, got 2 lines")
}

func Test_err_LineMismatch(t *testing.T) {
	mockT := &MockT{t: t}
	RunSysTest(mockT, runSysTestsDir, "err_linemismatch", internal.ExecRootCmd, "0.0.1")
	require.True(t, mockT.failed)
	mockT.assertMsgsContains("Line mismatch in req.md at line 1:\nexpected: line 1\ngot: line 1+")
}

// MockT implements a subset of testing.T for controlled failure testing
type MockT struct {
	t        *testing.T
	failed   bool
	failMsg  string
	failMsgs []string
}

func (m *MockT) Helper() {
	m.t.Helper()
}

func (m *MockT) String() string {
	return fmt.Sprintf("MockT: failed=%v, failMsgs=%v", m.failed, m.failMsgs)
}

func (m *MockT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = m.failMsg + "\n" + fmt.Sprintf(format, args...)
	m.failMsgs = append(m.failMsgs, fmt.Sprintf(format, args...))
}

func (m *MockT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = m.failMsg + "\n" + fmt.Sprintf(format, args...)
	panic("Fatalf called: " + m.failMsg)
}

func (m *MockT) FailNow() {
	panic("FailNow called")
}

func (m *MockT) TempDir() string {
	return m.t.TempDir()
}

func (m *MockT) assertMsgsContains(msg string) {
	m.t.Helper()
	for _, failMsg := range m.failMsgs {
		if strings.Contains(failMsg, msg) {
			return
		}
	}
	m.t.Errorf("Expected error not found in stderr: %s", msg)
}
