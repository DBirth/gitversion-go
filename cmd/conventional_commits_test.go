package cmd

import (
	"bytes"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tagRepo(t *testing.T, r *git.Repository, tagName string) {
	head, err := r.Head()
	require.NoError(t, err)

	_, err = r.CreateTag(tagName, head.Hash(), nil)
	require.NoError(t, err)
}

func TestRunCalculate_Conventional_Feat(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Initial commit and tag
	commitFile(t, w, "initial.txt", "initial commit")
	tagRepo(t, r, "1.0.0")

	// Feature commit
	commitFile(t, w, "feature.txt", "feat: a new feature")

	withMockGit(t, r, func() {
		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.1.0")
	})
}

func TestRunCalculate_Conventional_Fix(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Initial commit and tag
	commitFile(t, w, "initial.txt", "initial commit")
	tagRepo(t, r, "1.0.0")

	// Fix commit
	commitFile(t, w, "fix.txt", "fix: a bug fix")

	withMockGit(t, r, func() {
		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1")
	})
}

func TestRunCalculate_Conventional_BreakingChange(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Initial commit and tag
	commitFile(t, w, "initial.txt", "initial commit")
	tagRepo(t, r, "1.0.0")

	// Breaking change commit
	commitFile(t, w, "breaking.txt", "feat!: a breaking change")

	withMockGit(t, r, func() {
		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 2.0.0")
	})
}

func TestRunCalculate_Conventional_BreakingChangeBody(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Initial commit and tag
	commitFile(t, w, "initial.txt", "initial commit")
	tagRepo(t, r, "1.0.0")

	// Breaking change commit
	commitFile(t, w, "breaking.txt", "feat: a new feature\n\nBREAKING CHANGE: this breaks everything")

	withMockGit(t, r, func() {
		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 2.0.0")
	})
}
