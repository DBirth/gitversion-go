package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseBranch(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.checkout("release/1.0.0")
	repo.writeFile("GitVersion.yml", `
branches:
  release/*:
    mode: semver-from-branch
    tag: beta
`)
	repo.writeFile("release-file.txt", "update readme")
	repo.commit("docs: update readme")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, strings.Contains(string(output), "1.0.0-beta.1"), "output was: %s", string(output))
}
