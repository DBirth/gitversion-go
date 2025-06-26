package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevelopBranch(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.checkout("develop")
	repo.writeFile("GitVersion.yml", `
branches:
  develop:
    tag: alpha
`)
	repo.writeFile("develop-file.txt", "a new feature")
	repo.commit("feat: a new feature")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, strings.Contains(string(output), "1.1.0-alpha.1"), "output was: %s", string(output))
}
