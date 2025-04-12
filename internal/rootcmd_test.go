// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// skipIfNoRootCmdMarker skips the test if no .rootcmdYYMMDD file exists
func skipIfNoRootCmdMarker(t *testing.T) {
	pattern := fmt.Sprintf(".rootcmd%s", time.Now().Format("060102"))
	matches, err := filepath.Glob(".rootcmd*")
	if err != nil || len(matches) == 0 {
		t.Skipf("skipping test, no %s or other .rootcmd* file found", pattern)
	}
}

func Test_RootCmd_Draft(t *testing.T) {
	skipIfNoRootCmdMarker(t)

	require := require.New(t)

	err := ExecRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", "C:/workspaces/work/voedger-internals", "C:/workspaces/work/voedger"}, "0.0.1")
	require.Nil(err)
}

func Test_RootCmd_RelativePathsDry(t *testing.T) {
	skipIfNoRootCmdMarker(t)

	err := ExecRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", "../../voedger-internals", "../../voedger"}, "0.0.1")
	require.Nil(t, err)
}

// CAUTION: This test will apply changes to the files
func Test_RootCmd__Draft_RelativePathsApply(t *testing.T) {
	skipIfNoRootCmdMarker(t)

	// err := os.Chdir("C:/workspaces/work/voedger-internals/reqman")
	// require.Nil(t, err)
	err := ExecRootCmd([]string{"reqmd", "-v", "trace", "../../voedger-internals", "../../voedger"}, "0.0.1")
	require.Nil(t, err)
}
