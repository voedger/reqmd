// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package systrun

import "testing"

func Test_extractGoldenEmbedding(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		lines    []string
		expected []string
	}{
		{
			name:    "Basic line replacement",
			relPath: "test.md",
			lines: []string{
				"original content",
				"g: line: replaced content",
			},
			expected: []string{
				"replaced content",
				"g: line: replaced content",
			},
		},
		{
			name:    "Line removal directive",
			relPath: "test.md",
			lines: []string{
				"line to remove",
				"g: line- ",
				"another line",
			},
			expected: []string{
				"g: line- ",
				"another line",
			},
		},
		{
			name:    "Line addition directive",
			relPath: "test.md",
			lines: []string{
				"first line",
				"g: line+: new line after first",
				"last line",
			},
			expected: []string{
				"first line",
				"new line after first",
				"g: line+: new line after first",
				"last line",
			},
		},
		{
			name:    "Line at beginning directive",
			relPath: "test.md",
			lines: []string{
				"original content",
				"g: line1: first line",
				"g: line1: very first line",
			},
			expected: []string{
				"first line",
				"very first line",
				"original content",
				"g: line1: first line",
				"g: line1: very first line",
			},
		},
		{
			name:    "Line at end directive",
			relPath: "test.md",
			lines: []string{
				"original content",
				"g: line>>: last line",
				"g: line>>: very last line",
			},
			expected: []string{
				"original content",
				"g: line>>: last line",
				"g: line>>: very last line",
				"last line",
				"very last line",
			},
		},
		{
			name:    "Combination of directives",
			relPath: "test.md",
			lines: []string{
				"keep this line",
				"replace this line",
				"g: line: with this content",
				"remove this line",
				"g: line- ",
				"keep this line too",
				"g: line+: add this line after",
				"g: line1: add at beginning",
				"g: line>>: add at end",
			},
			expected: []string{
				"add at beginning",
				"keep this line",
				"with this content",
				"g: line: with this content",
				"g: line- ",
				"keep this line too",
				"add this line after",
				"g: line+: add this line after",
				"g: line1: add at beginning",
				"g: line>>: add at end",
				"add at end",
			},
		},
		{
			name:    "Multiple consecutive directives",
			relPath: "test.md",
			lines: []string{
				"original line",
				"g: line: replacement 1",
				"g: line: replacement 2",
				"normal line",
			},
			expected: []string{
				"replacement 2",
				"g: line: replacement 1",
				"g: line: replacement 2",
				"normal line",
			},
		},
		{
			name:    "Preserve comments not matching directives",
			relPath: "test.md",
			lines: []string{
				"original line",
				"g: this is a regular comment",
				"g: line+: add after original",
			},
			expected: []string{
				"original line",
				"add after original",
				"g: this is a regular comment",
				"g: line+: add after original",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyGoldenAnnotations(tt.lines)

			if len(result) != len(tt.expected) {
				t.Errorf("Length mismatch: got %d lines, want %d lines", len(result), len(tt.expected))
				t.Logf("Got: %v", result)
				t.Logf("Want: %v", tt.expected)
				return
			}

			for i := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing line at index %d, expected: %q", i, tt.expected[i])
					continue
				}
				if result[i] != tt.expected[i] {
					t.Errorf("Line %d mismatch:\ngot:  %q\nwant: %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
