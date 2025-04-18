// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package experiments

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/voedger/reqmd/internal"
)

// skipIfNoRootCmdMarker skips the test if no .expYYYYMMDD file exists (e.g. .exp2025041)
func skipIfNoRootCmdMarker(t *testing.T) {
	pattern := fmt.Sprintf(".exp%s*", time.Now().Format("20060102"))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		t.Skipf("skipping test, no %s file found", pattern)
	}
}

func Test_RootCmd_Draft(t *testing.T) {
	skipIfNoRootCmdMarker(t)

	require := require.New(t)

	err := internal.ExecRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", ".bugC:/workspaces/work/voedger-internals", "C:/workspaces/work/voedger"}, "0.0.1")
	require.Nil(err)
}

func Test_RootCmd_LocalVoedger_Dry(t *testing.T) {
	skipIfNoRootCmdMarker(t)
	err := internal.ExecRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", "../../../voedger-internals", "../../../voedger-internals/reqman/.work/repos/voedger"}, "0.0.1")
	require.Nil(t, err)
}

// CAUTION: This test will apply changes to the files
func Test_RootCmd_LocalVoedger(t *testing.T) {
	skipIfNoRootCmdMarker(t)
	err := internal.ExecRootCmd([]string{"reqmd", "-v", "trace", "../../../voedger-internals", "../../../voedger-internals/reqman/.work/repos/voedger"}, "0.0.1")
	require.Nil(t, err)
}

func Test_RootCmd_Data(t *testing.T) {
	skipIfNoRootCmdMarker(t)

	err := internal.ExecRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", ".data/voedger-internals/server/apiv2", "../../../voedger"}, "0.0.1")
	require.Nil(t, err)
}
