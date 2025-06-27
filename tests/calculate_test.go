// Package tests contains integration tests for the GitVersion CLI.
package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchBump(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.writeFile("another-file.txt", "another commit")
	repo.commit("chore: another commit")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "1.0.1")
}

func TestMinorBump(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.writeFile("GitVersion.yml", `minor-version-bump-message: '\+semver:\s?(minor)'`)
	repo.writeFile("another-file.txt", "another commit")
	repo.commit("+semver: minor")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "1.1.0")
}

func TestMajorBump(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.writeFile("GitVersion.yml", `major-version-bump-message: '\+semver:\s?(major)'`)
	repo.writeFile("another-file.txt", "another commit")
	repo.commit("+semver: major")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "2.0.0")
}

func TestFeatureBranch(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.checkout("feature/new-stuff")
	repo.writeFile("GitVersion.yml", `
branches:
  feature/*:
    tag: use-branch-name
`)
	repo.writeFile("feature-file.txt", "new feature on a branch")
	repo.commit("feat: new feature on a branch")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, strings.Contains(string(output), "1.1.0-feature-new-stuff.1"), "output was: %s", string(output))
}

func TestTagPrefix(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("release-1.0.0", initialCommit)

	repo.writeFile("GitVersion.yml", `tag-prefix: 'release-'`)
	repo.writeFile("another-file.txt", "another commit")
	repo.commit("chore: another commit")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "1.0.1")
}

func TestNoBumpMessage(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.writeFile("another-file.txt", "another commit")
	repo.commit("feat: another commit +semver: none")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "1.0.0")
}

func TestIgnoreSha(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.writeFile("another-file.txt", "another commit")
	ignoredCommit := repo.commit("feat: another commit")

	repo.writeFile("GitVersion.yml", "ignore:\n  - "+ignoredCommit.String())

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.Contains(t, string(output), "1.0.0")
}

func TestIncrementSetting(t *testing.T) {
	t.Run("BranchIncrementMinor", func(t *testing.T) {
		repo := newTestRepo(t)
		repo.writeFile("README.md", "initial commit")
		initialCommit := repo.commit("initial commit")
		repo.tag("v1.0.0", initialCommit)

		repo.checkout("feature/new-stuff")
		repo.writeFile("GitVersion.yml", `
branches:
  feature/*:
    increment: Minor
`)
		repo.writeFile("feature-file.txt", "new feature on a branch")
		repo.commit("chore: new feature on a branch")

		cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))

		assert.Contains(t, string(output), "1.1.0")
	})

	t.Run("BranchIncrementInherit", func(t *testing.T) {
		repo := newTestRepo(t)
		repo.writeFile("README.md", "initial commit")
		initialCommit := repo.commit("initial commit")
		repo.tag("v1.0.0", initialCommit)

		repo.checkout("feature/new-stuff")
		repo.writeFile("GitVersion.yml", `
branches:
  feature/*:
    increment: Inherit
`)
		repo.writeFile("feature-file.txt", "new feature on a branch")
		repo.commit("fix: a bug")

		cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))

		assert.Contains(t, string(output), "1.0.1")
	})

	t.Run("GlobalIncrementMajor", func(t *testing.T) {
		repo := newTestRepo(t)
		repo.writeFile("README.md", "initial commit")
		initialCommit := repo.commit("initial commit")
		repo.tag("v1.0.0", initialCommit)

		repo.writeFile("GitVersion.yml", `increment: Major`)
		repo.writeFile("another-file.txt", "another commit")
		repo.commit("chore: another commit")

		cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))

		assert.Contains(t, string(output), "2.0.0")
	})

	t.Run("BranchIncrementOverridesGlobal", func(t *testing.T) {
		repo := newTestRepo(t)
		repo.writeFile("README.md", "initial commit")
		initialCommit := repo.commit("initial commit")
		repo.tag("v1.0.0", initialCommit)

		repo.checkout("feature/new-stuff")
		repo.writeFile("GitVersion.yml", `
increment: Major
branches:
  feature/*:
    increment: Patch
`)
		repo.writeFile("feature-file.txt", "new feature on a branch")
		repo.commit("chore: new feature on a branch")

		cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, string(output))

		assert.Contains(t, string(output), "1.0.1")
	})
}

func TestFeatureBranch_MultipleCommits(t *testing.T) {
	repo := newTestRepo(t)
	repo.writeFile("README.md", "initial commit")
	initialCommit := repo.commit("initial commit")
	repo.tag("v1.0.0", initialCommit)

	repo.checkout("feature/new-stuff")
	repo.writeFile("GitVersion.yml", `
branches:
  feature/*:
    tag: use-branch-name
`)
	repo.writeFile("feature-file.txt", "new feature on a branch")
	repo.commit("feat: new feature on a branch")
	repo.writeFile("another-file.txt", "another commit")
	repo.commit("feat: another feature")

	cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, strings.Contains(string(output), "1.1.0-feature-new-stuff.2"), "output was: %s", string(output))
}
