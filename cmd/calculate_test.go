package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (*MemFs, *git.Repository, *git.Worktree, string) {
	tfs := NewMemFs()
	storage := memory.NewStorage()
	worktreeFs := memfs.New()

	r, err := git.Init(storage, worktreeFs)
	require.NoError(t, err)

	w, err := r.Worktree()
	require.NoError(t, err)

	// Create and add a file for the initial commit
	file, err := worktreeFs.Create("README.md")
	require.NoError(t, err)
	file.Close()
	_, err = w.Add("README.md")
	require.NoError(t, err)

	// Initial commit
	initialCommit, err := w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	require.NoError(t, err)

	// Tag initial commit as v1.0.0
	_, err = r.CreateTag("v1.0.0", initialCommit, nil)
	require.NoError(t, err)

	const repoPath = "/test-repo"
	tfs.Mkdir(repoPath, 0755)

	return tfs, r, w, repoPath
}

func withMockGit(t *testing.T, r *git.Repository, testFunc func()) {
	originalPlainOpen := gitPlainOpen
	gitPlainOpen = func(path string) (*git.Repository, error) {
		return r, nil
	}
	defer func() { gitPlainOpen = originalPlainOpen }()
	testFunc()
}

func commitFile(t *testing.T, w *git.Worktree, filename, msg string) {
	worktreeFs := w.Filesystem
	file, err := worktreeFs.Create(filename)
	require.NoError(t, err)
	file.Close()
	_, err = w.Add(filename)
	require.NoError(t, err)

	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	require.NoError(t, err)
}

func TestRunCalculate_PatchBump(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	withMockGit(t, r, func() {
		commitFile(t, w, "another-file.txt", "chore: another commit")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1")
	})
}

func TestRunCalculate_MinorBump(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Create a GitVersion.yml file for the test
	configContent := `minor-version-bump-message: '\+semver:\s?(minor)'`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "another-file.txt", "+semver: minor")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.1.0")
	})
}

func TestRunCalculate_MajorBump(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	// Create a GitVersion.yml file for the test
	configContent := `major-version-bump-message: '\+semver:\s?(major)'`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "another-file.txt", "+semver: major")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 2.0.0")
	})
}

func TestRunCalculate_FeatureBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	configContent := `
branches:
  feature:
    tag: use-branch-name
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Checkout a feature branch
	branchName := "feature/new-stuff"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "feature-file.txt", "chore: new feature on a branch")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1-feature-new-stuff.1")
	})
}

func TestRunCalculate_FeatureBranch_MultipleCommits(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)

	configContent := `
branches:
  feature:
    tag: use-branch-name
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := afero.WriteFile(fs, configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Checkout a feature branch
	branchName := "feature/new-stuff"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(t, err)

	withMockGit(t, r, func() {
		commitFile(t, w, "feature-file.txt", "chore: new feature on a branch")
		commitFile(t, w, "feature-file-2.txt", "chore: another feature on a branch")

		var out bytes.Buffer
		err := runCalculate(fs, &out, repoPath, "default")
		require.NoError(t, err)

		assert.Contains(t, out.String(), "Calculated next version: 1.0.1-feature-new-stuff.2")
	})
}
