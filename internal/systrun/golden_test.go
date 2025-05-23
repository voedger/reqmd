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
                ga("replace replaced content"),
            },
            expected: []string{
                "replaced content",
                ga("replace replaced content"),
            },
        },
        {
            name:    "Line removal directive",
            relPath: "test.md",
            lines: []string{
                "line to remove",
                ga("delete "),
                "another line",
            },
            expected: []string{
                ga("delete "),
                "another line",
            },
        },
        {
            name:    "Line addition directive",
            relPath: "test.md",
            lines: []string{
                "first line",
                ga("insert new line after first"),
                "last line",
            },
            expected: []string{
                "first line",
                "new line after first",
                ga("insert new line after first"),
                "last line",
            },
        },
        {
            name:    "Line at beginning directive",
            relPath: "test.md",
            lines: []string{
                "original content",
                ga("firstline first line"),
                ga("firstline very first line"),
            },
            expected: []string{
                "first line",
                "very first line",
                "original content",
                ga("firstline first line"),
                ga("firstline very first line"),
            },
        },
        {
            name:    "Line at end directive",
            relPath: "test.md",
            lines: []string{
                "original content",
                ga("append last line"),
                ga("append very last line"),
            },
            expected: []string{
                "original content",
                ga("append last line"),
                ga("append very last line"),
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
                ga("replace with this content"),
                "remove this line",
                ga("delete "),
                "keep this line too",
                ga("insert add this line after"),
                ga("firstline add at beginning"),
                ga("append add at end"),
            },
            expected: []string{
                "add at beginning",
                "keep this line",
                "with this content",
                ga("replace with this content"),
                ga("delete "),
                "keep this line too",
                "add this line after",
                ga("insert add this line after"),
                ga("firstline add at beginning"),
                ga("append add at end"),
                "add at end",
            },
        },
        {
            name:    "Multiple consecutive directives",
            relPath: "test.md",
            lines: []string{
                "original line",
                ga("replace replacement 1"),
                ga("replace replacement 2"),
                "normal line",
            },
            expected: []string{
                "replacement 2",
                ga("replace replacement 1"),
                ga("replace replacement 2"),
                "normal line",
            },
        },
        {
            name:    "Delete last line",
            relPath: "test.md",
            lines: []string{
                ga("deletelast"),
                "not last line",
                "last line",
            },
            expected: []string{
                ga("deletelast"),
                "not last line",
            },
        },
        {
            name:    "Delete and append multiple times",
            relPath: "test.md",
            lines: []string{
                ga("deletelast"),
                ga("deletelast"),
                ga("append appendedline1"),
                ga("append appendedline2"),
                "line1",
                "line2",
            },
            expected: []string{
                ga("deletelast"),
                ga("deletelast"),
                ga("append appendedline1"),
                ga("append appendedline2"),
                "appendedline1",
                "appendedline2",
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
