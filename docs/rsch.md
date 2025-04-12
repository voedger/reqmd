# Short research/study notes

## Are emojis searchable?

- 📌
  - gitbook: yes
  - VS Code: yes
  - github: no
- 🎯
  - gitbook: yes
  - VS Code: yes
  - github: ???

## go-git hash performance

```go
// C:/workspaces/work/voedger/pkg/istructsmem
// 37902	     30670 ns/op	       127.0 files	   24245 B/op	     642 allocs/op
//
// C:/workspaces/work/voedger
// 3440	    	301540 ns/op	      3590 files	  227868 B/op	    6194 allocs/op
func BenchmarkHash(b *testing.B) {

	repoRoot := "C:/workspaces/work/voedger"
	repo, err := git.PlainOpen("C:/workspaces/work/voedger")
	require.NoError(b, err)

	ref, err := repo.Head()
	require.NoError(b, err)

	commit, err := repo.CommitObject(ref.Hash())
	require.NoError(b, err)


	files, err := getFiles("C:/workspaces/work/voedger/pkg/istructsmem")
	require.NoError(b, err)

	// paths relative to the repo root
	for i, file := range files {
		relPath, err := filepath.Rel(repoRoot, file)
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		require.NoError(b, err)
		files[i] = relPath
	}

	tree, err := commit.Tree()
	_ = tree
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, file := range files {
			if !strings.HasSuffix(file, ".go") {
				continue
			}
			entry, err := tree.FindEntry(file)
			if err != nil {
				b.Fatal(err, file)
			}
			hash := entry.Hash.String()
			if len(hash) == 0 {
				panic("hash is empty")
			}

		}
	}
	b.ReportMetric(float64(len(files)), "files")
}

func getFiles(directory string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {

		if err != nil {

			return err

		}

		// Normalize path to avoid mismatches
		normalizedPath := strings.ReplaceAll(path, "\\", "/")

		// Add only files (not directories)
		if !d.IsDir() {
			files = append(files, normalizedPath)
		}

		return nil
	})

	if err != nil {

		return nil, err

	}

	return files, nil
}
```
