package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"gitversion-go/internal/pkg/fs"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) (fs.Filesystem, *git.Repository, *git.Worktree, string) {
	storage := memory.NewStorage()
	worktreeFs := memfs.New()
	fs := fs.NewBillyWrappedFs(worktreeFs)

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

	return fs, r, w, "/"
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
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()

	commitFile(t, w, "another-file.txt", "chore: another commit")

	var out bytes.Buffer
	err := runCalculate(fs, &out, repoPath, "default")
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Calculated next version: 1.0.1")
}

func TestRunCalculate_MinorBump(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()

	// Create a GitVersion.yml file for the test
	configContent := `minor-version-bump-message: '\+semver:\s?(minor)'`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := fs.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	commitFile(t, w, "another-file.txt", "+semver: minor")

	var out bytes.Buffer
	err = runCalculate(fs, &out, repoPath, "default")
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Calculated next version: 1.1.0")
}

func TestRunCalculate_MajorBump(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()

	// Create a GitVersion.yml file for the test
	configContent := `major-version-bump-message: '\+semver:\s?(major)'`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := fs.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	commitFile(t, w, "another-file.txt", "+semver: major")

	var out bytes.Buffer
	err = runCalculate(fs, &out, repoPath, "default")
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Calculated next version: 2.0.0")
}

func TestRunCalculate_FeatureBranch(t *testing.T) {
	fs, r, w, repoPath := setupTestRepo(t)
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()

	// Checkout a feature branch
	branchName := "feature/new-stuff"
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err := w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(t, err)

	configContent := `
branches:
  feature/*:
    tag: use-branch-name
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err = fs.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	commitFile(t, w, "feature-file.txt", "feat: new feature on a branch")

	var out bytes.Buffer
	err = runCalculate(fs, &out, repoPath, "default")
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Calculated next version: 1.1.0-feature-new-stuff.1")
}

func TestRunCalculate_FeatureBranch_MultipleCommits(t *testing.T) {
	fs, r, _, repoPath := setupTestRepo(t)
	gitPlainOpen = func(path string) (*git.Repository, error) { return r, nil }
	defer func() { gitPlainOpen = git.PlainOpen }()
	configContent := `
branches:
  feature/*:
    tag: use-branch-name
`
	configPath := filepath.Join(repoPath, "GitVersion.yml")
	err := fs.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
}
