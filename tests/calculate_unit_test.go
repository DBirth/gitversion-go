// Package tests contains integration and unit tests for the GitVersion CLI and core logic.
package tests

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

	"gitversion-go/internal/app"
)

func TestRunCalculate_Unit(t *testing.T) {
	testCases := []struct {
		name           string
		config         string
		setupRepo      func(t *testing.T, r *git.Repository, worktree *git.Worktree)
		expectedOutput string
		expectedErr    string
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
				commitUnit(t, worktree, "feat: Initial commit")
				tagUnit(t, r, "v1.0.0")
				commitUnit(t, worktree, "feat!: Breaking change")
			},
			expectedOutput: "Calculated next version: 2.0.0\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fs.NewOsFs()
			tempDir := t.TempDir()
			r, err := git.PlainInit(tempDir, false)
			require.NoError(t, err)
			worktree, err := r.Worktree()
			require.NoError(t, err)

			configPath := filepath.Join(tempDir, "GitVersion.yml")
			err = fsys.WriteFile(configPath, []byte(tc.config), 0644)
			require.NoError(t, err)

			if tc.setupRepo != nil {
				tc.setupRepo(t, r, worktree)
			}

			var out bytes.Buffer
			err = app.RunCalculate(fsys, &out, tempDir, "default")

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, out.String())
			}
		})
	}
}

func commitUnit(t *testing.T, w *git.Worktree, msg string) {
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

func tagUnit(t *testing.T, r *git.Repository, tagName string) {
	t.Helper()
	head, err := r.Head()
	require.NoError(t, err)
	_, err = r.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: tagName,
		Tagger: &object.Signature{
			Name:  "TestTagger",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
}
