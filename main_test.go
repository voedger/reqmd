package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MainDraft(t *testing.T) {
	require := require.New(t)

	err := execRootCmd([]string{"reqmd", "-v", "trace", "C:/workspaces/work/voedger-internals", "C:/workspaces/work/voedger"}, "0.0.1")
	require.Nil(err)

	require.NotNil(t)
}
