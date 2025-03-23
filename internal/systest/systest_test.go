package systest

import (
	"fmt"
	"os"
	"testing"
)

const testsDir = "testdata"

func Test_noreqs(t *testing.T) {
	RunSysTest(t, testsDir, "noreqs", []string{"trace"}, "0.0.1")
}

func Test_err_undetected(t *testing.T) {
	// https://claude.ai/chat/bfbb2bec-ab1f-42b2-bb87-2447989fe68f

	mockT := &MockT{}
	defer mockT.Cleanup()

	RunSysTest(t, testsDir, "err_undetected", []string{"trace"}, "0.0.1")
}

// MockT implements a subset of testing.T for controlled failure testing
type MockT struct {
	failed   bool
	failMsg  string
	tempDirs []string
}

func (m *MockT) Errorf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = fmt.Sprintf(format, args...)
}

func (m *MockT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.failMsg = fmt.Sprintf(format, args...)
}

func (m *MockT) FailNow() {
	m.failed = true
}

func (m *MockT) TempDir() string {
	dir, err := os.MkdirTemp("", "systest-")
	if err != nil {
		m.Fatalf("Failed to create temp dir: %v", err)
		return ""
	}
	m.tempDirs = append(m.tempDirs, dir)
	return dir
}

func (m *MockT) Cleanup() {
	for _, dir := range m.tempDirs {
		os.RemoveAll(dir)
	}
}
