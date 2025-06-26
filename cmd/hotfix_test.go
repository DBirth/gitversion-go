package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCalculate_HotfixBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	configContent := `
branches:
  hotfix:
    mode: semver-from-branch
    tag: beta
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Checkout a hotfix branch
	branchName := "hotfix/1.0.1"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "hotfix-file.txt", "fix: a critical hotfix")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1-beta.1")
	})
}
