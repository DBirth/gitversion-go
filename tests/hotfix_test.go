package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotfixBranch(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.checkout("hotfix/1.0.1")
	repo.writeFile("GitVersion.yml", `
branches:
  hotfix/*:
    mode: semver-from-branch
    tag: beta
`)
	repo.writeFile("hotfix-file.txt", "a hotfix")
	repo.commit("fix: a hotfix")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, strings.Contains(string(output), "1.0.1-beta.1"), "output was: %s", string(output))
}
