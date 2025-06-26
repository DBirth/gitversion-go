package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCalculate_DevelopBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Checkout develop branch
	branchName := "develop"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err := w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(t, err)

	configContent := `
branches:
  develop:
    tag: alpha
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err = fs.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "develop-file.txt", "feat: a new feature")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.1.0-alpha.1")
	})
}
