// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortCoverers(t *testing.T) {
	tests := []struct {
		name     string
		coverers []Coverer
		want     []Coverer
	}{
		{
			name: "sort by CoverageType",
			coverers: []Coverer{
				{CoverageLabel: "file.go:1:test", CoverageURL: "url1"},
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "file.go:1:test", CoverageURL: "url1"},
			},
		},
		{
			name: "sort by FilePath",
			coverers: []Coverer{
				{CoverageLabel: "z.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "a.go:1:impl", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "a.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "z.go:1:impl", CoverageURL: "url1"},
			},
		},
		{
			name: "sort by Number",
			coverers: []Coverer{
				{CoverageLabel: "file.go:20:impl", CoverageURL: "url1"},
				{CoverageLabel: "file.go:3:impl", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "file.go:3:impl", CoverageURL: "url1"},
				{CoverageLabel: "file.go:20:impl", CoverageURL: "url1"},
			},
		},
		{
			name: "sort by CoverageURL",
			coverers: []Coverer{
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url2"},
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "file.go:1:impl", CoverageURL: "url2"},
			},
		},
		{
			name: "complex sort",
			coverers: []Coverer{
				{CoverageLabel: "b.go:12:test", CoverageURL: "url2"},
				{CoverageLabel: "b.go:1:impl", CoverageURL: "url2"},
				{CoverageLabel: "a.go:3:impl", CoverageURL: "url1"},
				{CoverageLabel: "a.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "b.go:2:test", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "a.go:1:impl", CoverageURL: "url1"},
				{CoverageLabel: "a.go:3:impl", CoverageURL: "url1"},
				{CoverageLabel: "b.go:1:impl", CoverageURL: "url2"},
				{CoverageLabel: "b.go:2:test", CoverageURL: "url1"},
				{CoverageLabel: "b.go:12:test", CoverageURL: "url2"},
			},
		},
		{
			name: "invalid format handling",
			coverers: []Coverer{
				{CoverageLabel: "invalid", CoverageURL: "url2"},
				{CoverageLabel: "also:invalid", CoverageURL: "url1"},
			},
			want: []Coverer{
				{CoverageLabel: "also:invalid", CoverageURL: "url1"},
				{CoverageLabel: "invalid", CoverageURL: "url2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortCoverers(tt.coverers)
			assert.Equal(t, tt.want, tt.coverers)
		})
	}
}

func TestFormatCoverageFootnote(t *testing.T) {
	tests := []struct {
		name     string
		footnote *CoverageFootnote
		want     string
	}{
		{
			name: "no coverers",
			footnote: &CoverageFootnote{
				PackageID:          "pkg1",
				CoverageFootnoteId: "001",
				RequirementName:    "REQ001",
			},
			want: "[^001]: `[~pkg1/REQ001~impl]`",
		},
		{
			name: "with sorted coverers",
			footnote: &CoverageFootnote{
				CoverageFootnoteId: "002",
				PackageID:          "pkg2",
				RequirementName:    "REQ001",
				Coverers: []Coverer{
					{CoverageLabel: "b.go:1:test", CoverageURL: "url2"},
					{CoverageLabel: "a.go:1:impl", CoverageURL: "url1"},
				},
			},
			want: "[^002]: `[~pkg2/REQ001~impl]` [a.go:1:impl](url1), [b.go:1:test](url2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCoverageFootnote(tt.footnote)
			assert.Equal(t, tt.want, got)
		})
	}
}
