package systest

import (
	"embed"
	"testing"
)

//go:embed testdata/*
var testdata embed.FS

func Test_noreqs(t *testing.T) {
	RunSysTest(t, testdata, "noreqs", []string{"trace"}, "0.0.1")
}

func Test_err_undetected(t *testing.T) {
	// https://claude.ai/chat/bfbb2bec-ab1f-42b2-bb87-2447989fe68f
	RunSysTest(t, testdata, "err_undetected", []string{"trace"}, "0.0.1")
}

// type SysTestSuite struct {
// 	suite.Suite
// 	mockT *testing.T
// }

// func (s *SysTestSuite) Cleanup(_ func()) {
// }

// func (s *SysTestSuite) TestRunSysTestFailure() {
// 	// Create a new test context that won't fail the actual test
// 	mockT := new(SysTestSuite)

// 	RunSysTest(mockT, testdata, "err_undetected", []string{"trace"}, "0.0.1")

// 	// If mockT didn't fail as expected, fail the actual test
// 	s.True(mockT.Failed(), "Expected RunSysTest to fail, but it did not")
// }

// func TestSysTestSuite(t *testing.T) {
// 	suite.Run(t, new(SysTestSuite))
// }
