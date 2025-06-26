package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCalculate_MainBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	configContent := `
branches:
  main:
    tag: ""
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "main-file.txt", "fix: a fix on main")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1")
	})
}
