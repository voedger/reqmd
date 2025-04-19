// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0

package hvgen

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"time"

	"github.com/voedger/reqmd/internal"
)

// Config defines the parameters for high-volume test generation
type Config struct {
	NumReqSites        int
	MaxSitesPerPackage int
	MaxTagsPerSite     int
	MaxSitesPerFile    int
	MaxTagsPerFile     int
	MaxTreeDepth       int
	SrcToMdRatio       int
}

// DefaultConfig provides sensible defaults for Config
func DefaultConfig() Config {
	return Config{
		NumReqSites:        1000,
		MaxSitesPerPackage: 5,
		MaxTagsPerSite:     2,
		MaxSitesPerFile:    3,
		MaxTagsPerFile:     3,
		MaxTreeDepth:       4,
		SrcToMdRatio:       5,
	}
}

// HVGenerator generates a test file structure with configurable parameters
func HVGenerator(cfg Config) ([]internal.FileStructure, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate folder names
	folderNames := generateFolderNames(cfg.MaxTreeDepth)

	// Generate requirement IDs
	reqIds := generateRequirementIds(cfg.NumReqSites, cfg.MaxSitesPerPackage)

	// Group requirement IDs by package for file distribution
	reqIdPerFile := groupRequirementIdsPerFile(reqIds, cfg.MaxSitesPerFile)

	// Generate coverage tags
	ctags, _ := generateCoverageTags(reqIds, cfg.MaxTagsPerSite)

	// Group coverage tags per file
	ctagPerFile := groupCoverageTagsPerFile(r, ctags, cfg.MaxTagsPerFile)

	// Generate file structures
	fileStructs := generateFileStructures(r, reqIdPerFile, ctagPerFile, folderNames, cfg.SrcToMdRatio)

	return fileStructs, nil
}

// generateFolderNames generates a slice of folder names
// in the format ["f1", "f2", "f3"...] with length equal to maxDepth
func generateFolderNames(maxDepth int) []string {
	folderNames := make([]string, maxDepth)
	for i := range folderNames {
		folderNames[i] = fmt.Sprintf("f%d", i+1)
	}
	return folderNames
}

// generateRequirementIds generates requirement IDs based on configuration
// Input: numReqSites, avgSitesPerPackage
// Output: []RequirementId sorted by PackageId
//
// Requirements are distributed across packages randomly with avgSitesPerPackage metric
// Each requirement has a unique name and is associated with a specific package ID.

func generateRequirementIds(numReqSites int, maxSitesPerPackage int) []internal.RequirementId {
	if numReqSites <= 0 || maxSitesPerPackage <= 0 {
		return []internal.RequirementId{}
	}

	// Calculate approximately how many packages we need
	numPackages := numReqSites * 2 / maxSitesPerPackage
	if numPackages < 1 {
		numPackages = 1
	}

	result := make([]internal.RequirementId, numReqSites)

	// Distribute requirements across packages
	for i := range numReqSites {
		packageIdx := rand.Intn(numPackages)

		// Create a requirement ID with a unique name
		reqId := internal.RequirementId{
			PackageId:       internal.PackageId(fmt.Sprintf("pkg%d", packageIdx+1)),
			RequirementName: internal.RequirementName(fmt.Sprintf("req%d", i+1)),
		}

		result[i] = reqId
	}

	// Sort by PackageId to group related requirements together
	sort.Slice(result, func(i, j int) bool {
		return string(result[i].PackageId) < string(result[j].PackageId)
	})

	return result
}

// groupRequirementIdsPerFile groups reqIds into files according to avgSitesPerFile parameter
// Output: [][]RequirementId where each element contains
// RequirementIds with the same PackageId
//
// This function distributes requirements to files while ensuring
// requirements with the same package ID stay together.
// Flow:
// - Initialize current group (cg): currentPackageId, cgNumReqs
// - Iterate over reqIds
//   - If reqIds.PackageId != cgPackageId or len(cg) >= cgNumReqs: flush cg to result and start new cg
func groupRequirementIdsPerFile(reqIds []internal.RequirementId, maxSitesPerFile int) [][]internal.RequirementId {
	if len(reqIds) == 0 || maxSitesPerFile <= 0 {
		return [][]internal.RequirementId{}
	}

	var result [][]internal.RequirementId

	var currentGroup []internal.RequirementId
	var currentPackageId internal.PackageId
	var currentGroupSize int

	initGroup := func(reqId internal.RequirementId) {
		currentPackageId = reqId.PackageId
		currentGroupSize = rand.Intn(maxSitesPerFile + 1)
		currentGroup = []internal.RequirementId{}
	}
	initGroup(reqIds[0])

	for _, reqId := range reqIds {
		// If we encounter a new package or reached max group size, flush the current group
		if reqId.PackageId != currentPackageId || len(currentGroup) >= currentGroupSize {
			if len(currentGroup) > 0 {
				result = append(result, currentGroup)
				currentGroup = nil
			}
			initGroup(reqId)
		}

		// Add the current reqId to the group
		currentGroup = append(currentGroup, reqId)
	}

	// Don't forget to add the last group
	if len(currentGroup) > 0 {
		result = append(result, currentGroup)
	}
	return result
}

// generateCoverageTags generates coverage tags for requirements
// Input: AvgTagsPerSite, reqIds
// Output: tags []CoverageTag, reqToTags map[RequirementId][]CoverageTags
//
// This function creates coverage tags for each requirement ID,
// with random coverage types ("impl" or "test") and line numbers.
// The number of tags per requirement is determined by maxTagsPerSite.
func generateCoverageTags(reqIds []internal.RequirementId, maxTagsPerSite int) ([]internal.CoverageTag, map[internal.RequirementId][]internal.CoverageTag) {
	coverageTypes := []string{"impl", "test"}
	var allTags []internal.CoverageTag
	reqToTags := make(map[internal.RequirementId][]internal.CoverageTag)

	for _, reqId := range reqIds {
		// Determine number of tags for this requirement
		numTags := rand.Intn(maxTagsPerSite + 1)

		// Generate tags for this requirement
		tags := make([]internal.CoverageTag, numTags)
		for i := range numTags {
			tag := internal.CoverageTag{
				RequirementId: reqId,
				CoverageType:  coverageTypes[rand.Intn(len(coverageTypes))],
				Line:          rand.Intn(100) + 1, // Line numbers from 1 to 100
			}
			tags[i] = tag
			allTags = append(allTags, tag)
		}
		reqToTags[reqId] = tags
	}

	return allTags, reqToTags
}

// groupCoverageTagsPerFile groups coverage tags into files
// Input: ctags, cfg.AvgTagsPerFile
// Output: [][]CoverageTags
//
// This function distributes coverage tags across files,
// attempting to keep tags for the same requirement together.
func groupCoverageTagsPerFile(r *rand.Rand, ctags []internal.CoverageTag, avgTagsPerFile int) [][]internal.CoverageTag {
	if len(ctags) == 0 || avgTagsPerFile <= 0 {
		return [][]internal.CoverageTag{}
	}

	// Calculate number of files needed
	numFiles := len(ctags) / avgTagsPerFile
	if numFiles < 1 {
		numFiles = 1
	}

	result := make([][]internal.CoverageTag, numFiles)

	// Create a map to group tags by requirement ID
	tagsByReqId := make(map[internal.RequirementId][]internal.CoverageTag)
	for _, tag := range ctags {
		tagsByReqId[tag.RequirementId] = append(tagsByReqId[tag.RequirementId], tag)
	}

	fileIdx := 0

	// Distribute tags to files, keeping tags with same requirement ID together
	for _, reqTags := range tagsByReqId {
		// Determine how many tags to add to this file
		count := avgTagsPerFile
		variation := r.Intn(3) - 1 // -1, 0, or 1
		count += variation
		if count <= 0 {
			count = 1
		}
		if count > len(reqTags) {
			count = len(reqTags)
		}

		// Add tags to the current file
		result[fileIdx] = append(result[fileIdx], reqTags[:count]...)

		fileIdx = (fileIdx + 1) % numFiles
	}

	return result
}

// generateFileStructures creates file structures based on grouped requirements and tags
// Input: reqIdPerFile, ctagPerFile, folderNames
// Output: []FileStructure
//
// This function creates FileStructure objects by generating:
// - Path: from random elements of folderNames
// - Type: Markdown or Source (ratio determined by srcToMdRatio)
// - PackageId: from the requirements assigned to the file
// - Requirements: from reqIdPerFile with unique line numbers
// - CoverageTags: from ctagPerFile with unique line numbers
func generateFileStructures(r *rand.Rand, reqIdPerFile [][]internal.RequirementId, ctagPerFile [][]internal.CoverageTag, folderNames []string, srcToMdRatio int) []internal.FileStructure {
	maxFiles := max(len(reqIdPerFile), len(ctagPerFile))
	result := make([]internal.FileStructure, maxFiles)

	for i := 0; i < maxFiles; i++ {
		fs := internal.FileStructure{}

		// Generate path with random folder depth
		depth := r.Intn(len(folderNames)) + 1
		pathParts := make([]string, depth)
		for j := 0; j < depth; j++ {
			pathParts[j] = folderNames[r.Intn(len(folderNames))]
		}

		// Determine file type (markdown or source)
		fileType := internal.FileTypeSource
		if r.Intn(srcToMdRatio+1) == 0 { // 1 in (srcToMdRatio+1) chance of being markdown
			fileType = internal.FileTypeMarkdown
			pathParts = append(pathParts, fmt.Sprintf("doc%d.md", i))
		} else {
			pathParts = append(pathParts, fmt.Sprintf("file%d.go", i))
		}

		fs.Path = filepath.Join(pathParts...)
		fs.Type = fileType

		// Add requirements if available for this file index
		if i < len(reqIdPerFile) && len(reqIdPerFile[i]) > 0 {
			fs.PackageId = reqIdPerFile[i][0].PackageId
			usedLines := make(map[int]bool)

			for _, reqId := range reqIdPerFile[i] {
				// Generate a unique line number
				line := r.Intn(100) + 1
				for usedLines[line] {
					line = r.Intn(100) + 1
				}
				usedLines[line] = true

				// Create requirement site
				reqSite := internal.RequirementSite{
					Line:               line,
					RequirementName:    reqId.RequirementName,
					CoverageFootnoteId: internal.CoverageFootnoteId(fmt.Sprintf("%d", len(fs.Requirements)+1)),
				}

				fs.Requirements = append(fs.Requirements, reqSite)
			}
		}

		// Add coverage tags if available for this file index
		if i < len(ctagPerFile) {
			usedLines := make(map[int]bool)

			for _, tag := range ctagPerFile[i] {
				// Generate a unique line number
				line := r.Intn(100) + 1
				for usedLines[line] {
					line = r.Intn(100) + 1
				}
				usedLines[line] = true

				tag.Line = line
				fs.CoverageTags = append(fs.CoverageTags, tag)
			}
		}

		// Set other fields
		fs.RepoRootFolderURL = "http://example.com/repo"
		fs.RelativePath = fs.Path
		fs.FileHash = fmt.Sprintf("hash%d", i)

		result[i] = fs
	}

	return result
}

// helper function for max of two ints (for Go <1.21)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
