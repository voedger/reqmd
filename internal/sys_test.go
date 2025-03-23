package internal

import (
	"path/filepath"
	"testing"

	"github.com/voedger/reqmd/internal/systest"
)

var sysTestsDir = filepath.Join("testdata", "systest")

func Test_systest_NoReqs(t *testing.T) {
	systest.RunSysTest(t, sysTestsDir, "noreqs", ExecRootCmd, []string{"trace"}, "0.0.1")
}
