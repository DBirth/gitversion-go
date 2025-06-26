package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

type testRepo struct {
	*git.Repository
	worktree *git.Worktree
	path     string
	t        *testing.T
}

func newTestRepo(t *testing.T) *testRepo {
	tempDir, err := os.MkdirTemp("", "gitversion-go-tests-repo")
	require.NoError(t, err)

	r, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	w, err := r.Worktree()
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return &testRepo{Repository: r, worktree: w, path: tempDir, t: t}
}

func (r *testRepo) commit(msg string) plumbing.Hash {
	commit, err := r.worktree.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	require.NoError(r.t, err)
	return commit
}

func (r *testRepo) tag(tag string, hash plumbing.Hash) {
	_, err := r.Repository.CreateTag(tag, hash, nil)
	require.NoError(r.t, err)
}

func (r *testRepo) checkout(branch string) {
	branchRef := plumbing.NewBranchReferenceName(branch)
	err := r.worktree.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: branchRef,
	})
	require.NoError(r.t, err)
}

func (r *testRepo) writeFile(filename, content string) {
	filePath := filepath.Join(r.path, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(r.t, err)

	_, err = r.worktree.Add(filename)
	require.NoError(r.t, err)
}
