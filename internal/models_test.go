package internal

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReqmdjson_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Reqmdjson
		wantErr bool
	}{
		{
			name: "valid json with multiple file hashes",
			json: `{
				"FileURL2FileHash": {
					"https://github.com/voedger/voedger/blob/main/pkg/api/handler.go": "979d75b2c7da961f94396ce2b286e7389eb73d75",
					"https://github.com/voedger/voedger/blob/main/pkg/api/handler_test.go": "845a23c8f9d6a8b7e9c2d4f5a6b7c8d9e0f1a2b3"
				}
			}`,
			want: Reqmdjson{
				FileURL2FileHash: map[string]string{
					"https://github.com/voedger/voedger/blob/main/pkg/api/handler.go":      "979d75b2c7da961f94396ce2b286e7389eb73d75",
					"https://github.com/voedger/voedger/blob/main/pkg/api/handler_test.go": "845a23c8f9d6a8b7e9c2d4f5a6b7c8d9e0f1a2b3",
				},
			},
		},
		{
			name: "empty file hashes",
			json: `{"FileURL2FileHash":{}}`,
			want: Reqmdjson{
				FileURL2FileHash: map[string]string{},
			},
		},
		{
			name:    "invalid json",
			json:    `{"FileURL2FileHash": not_valid_json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Reqmdjson
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.FileURL2FileHash, got.FileURL2FileHash)
		})
	}
}

func TestReqmdjson_MarshalJSON_sorted(t *testing.T) {
	input := Reqmdjson{
		FileURL2FileHash: map[string]string{
			// Deliberately not in lexical order
			"https://github.com/org/repo/blob/main/zzz/last.go":      "hash20",
			"https://github.com/org/repo/blob/main/src/app.go":       "hash10",
			"https://github.com/org/repo/blob/main/pkg/main.go":      "hash01",
			"https://github.com/org/repo/blob/main/test/b_test.go":   "hash15",
			"https://github.com/org/repo/blob/main/pkg/utils/io.go":  "hash03",
			"https://github.com/org/repo/blob/main/cmd/app/main.go":  "hash07",
			"https://github.com/org/repo/blob/main/internal/core.go": "hash08",
			"https://github.com/org/repo/blob/main/pkg/api/v1.go":    "hash02",
			"https://github.com/org/repo/blob/main/test/a_test.go":   "hash14",
			"https://github.com/org/repo/blob/main/docs/README.md":   "hash06",
			"https://github.com/org/repo/blob/main/src/lib.go":       "hash11",
			"https://github.com/org/repo/blob/main/pkg/db/sql.go":    "hash04",
			"https://github.com/org/repo/blob/main/src/types.go":     "hash13",
			"https://github.com/org/repo/blob/main/src/mock.go":      "hash12",
			"https://github.com/org/repo/blob/main/pkg/log/log.go":   "hash05",
			"https://github.com/org/repo/blob/main/test/e2e.go":      "hash16",
			"https://github.com/org/repo/blob/main/tools/gen.go":     "hash17",
			"https://github.com/org/repo/blob/main/ui/app.tsx":       "hash18",
			"https://github.com/org/repo/blob/main/web/index.ts":     "hash19",
			"https://github.com/org/repo/blob/main/cmd/cli.go":       "hash09",
		},
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	// Verify exact number of entries
	assert.Equal(t, 20, len(input.FileURL2FileHash), "should have exactly 20 entries")

	// Unmarshal to verify structure
	var output Reqmdjson
	err = json.Unmarshal(data, &output)
	require.NoError(t, err)

	// Verify content equality
	assert.Equal(t, input.FileURL2FileHash, output.FileURL2FileHash)

	// Extract and sort all keys
	keys := make([]string, 0, len(input.FileURL2FileHash))
	for k := range input.FileURL2FileHash {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Convert to string for easier substring search
	jsonStr := string(data)

	// Check that each key appears in the correct order in the JSON string
	for i := 0; i < len(keys)-1; i++ {
		currentKey := keys[i]
		nextKey := keys[i+1]
		currentIdx := strings.Index(jsonStr, currentKey)
		nextIdx := strings.Index(jsonStr, nextKey)
		assert.Less(t, currentIdx, nextIdx,
			"Key '%s' should appear before '%s' in JSON output", currentKey, nextKey)
	}
}
