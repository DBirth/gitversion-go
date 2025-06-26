package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCalculate_MainBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	configContent := `
branches:
  main:
    tag: beta
`
	err := fs.WriteFile("GitVersion.yml", []byte(configContent), 0644)
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "main-file.txt", "fix: a fix on main")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1")
	})
}
