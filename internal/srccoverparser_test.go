package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceParser_ParseSourceFile(t *testing.T) {
	// Test data file contains a line with: [~server.api.v2/Post.handler~test]
	testDataFile := filepath.Join("testdata", "srccoverparser-1.go")
	srcFile, syntaxErrors, err := ParseSourceFile(testDataFile)
	require.NoError(t, err)
	assert.Len(t, syntaxErrors, 0)

	// Verify file type and that at least one coverage tag is found
	assert.Equal(t, FileTypeSource, srcFile.Type)
	require.NotEmpty(t, srcFile.CoverageTags)

	{
		tag := srcFile.CoverageTags[0]
		assert.Equal(t, RequirementId("server.api.v2/Post.handler"), tag.RequirementId)
		assert.Equal(t, "impl", tag.CoverageType)
		// Adjust expected line number according to your test file content
		assert.Equal(t, 8, tag.Line)
	}

	{
		tag := srcFile.CoverageTags[1]
		assert.Equal(t, RequirementId("server.api.v2/Post.handler"), tag.RequirementId)
		assert.Equal(t, "test", tag.CoverageType)
		// Adjust expected line number according to your test file content
		assert.Equal(t, 14, tag.Line)
	}
}
