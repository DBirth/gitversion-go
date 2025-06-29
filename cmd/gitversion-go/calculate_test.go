package main

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitversion-go/internal/fs"
)

func TestRunCalculate(t *testing.T) {
	testCases := []struct {
		name               string
		config             string
		setupRepo          func(t *testing.T, r *git.Repository, worktree *git.Worktree) 
		expectedOutput     string
		expectedErr        string
	}{
		{
			name: "Simple major bump from commit message",
			config: `
next-version: 1.0.0
strategies:
  - find-latest-tag
  - increment-from-commits
`,
			setupRepo: func(t *testing.T, r *git.Repository, worktree *git.Worktree) {
				commit(t, worktree, "feat: Initial commit")
				tag(t, r, "v1.0.0")
				commit(t, worktree, "feat!: Breaking change")
			},
			expectedOutput: "Calculated next version: 2.0.0\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			fs := fs.NewOsFs()
			tempDir := t.TempDir()
			r, err := git.PlainInit(tempDir, false)
			require.NoError(t, err)
			worktree, err := r.Worktree()
			require.NoError(t, err)

			// Write config file
			configPath := filepath.Join(tempDir, "GitVersion.yml")
			err = fs.WriteFile(configPath, []byte(tc.config), 0644)
			require.NoError(t, err)

			// Setup repo state
			if tc.setupRepo != nil {
				tc.setupRepo(t, r, worktree)
			}

			// Execute
			var out bytes.Buffer
			err = runCalculate(fs, &out, tempDir, "default")

			// Assert
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, out.String())
			}
		})
	}
}

func commit(t *testing.T, w *git.Worktree, msg string) {
	t.Helper()
	filename := "dummy_file.txt"
	file, err := w.Filesystem.Create(filename)
	require.NoError(t, err)
	_, err = file.Write([]byte(msg))
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	_, err = w.Add(filename)
	require.NoError(t, err)

	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
}

func tag(t *testing.T, r *git.Repository, tagName string) {
	t.Helper()
	head, err := r.Head()
	require.NoError(t, err)
	_, err = r.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: tagName,
	})
	require.NoError(t, err)
}
