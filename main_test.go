// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MainDraft(t *testing.T) {
	require := require.New(t)

	err := execRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", "C:/workspaces/work/voedger-internals", "C:/workspaces/work/voedger"}, "0.0.1")
	require.Nil(err)

	require.NotNil(t)
}

func Test_MainDraft_RelativePathsDry(t *testing.T) {
	// err := os.Chdir("C:/workspaces/work/voedger-internals/reqman")
	// require.Nil(t, err)
	err := execRootCmd([]string{"reqmd", "-v", "trace", "--dry-run", "../voedger-internals", "../voedger"}, "0.0.1")
	require.Nil(t, err)
}

// CAUTION: This test will apply changes to the files
func Test_MainDraft_RelativePathsApply(t *testing.T) {
	// err := os.Chdir("C:/workspaces/work/voedger-internals/reqman")
	// require.Nil(t, err)
	err := execRootCmd([]string{"reqmd", "-v", "trace", "../voedger-internals", "../voedger"}, "0.0.1")
	require.Nil(t, err)
}
