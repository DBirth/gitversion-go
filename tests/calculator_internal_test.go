package tests

import (
	"gitversion-go/internal/gitversion"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestFindLatestVersionWithSourceBranches(t *testing.T) {
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Make a change before the second commit
	file2, err := fs.Create("CHANGELOG.md")
	assert.NoError(t, err)
	_, err = file2.Write([]byte("changelog content"))
	assert.NoError(t, err)
	err = file2.Close()
	assert.NoError(t, err)
	_, err = w.Add("CHANGELOG.md")
	assert.NoError(t, err)

	_, err = w.Commit("Second commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	_, err = r.CreateTag("v1.0.0", commit1, nil)
	assert.NoError(t, err)

	config := &gitversion.Config{
		Branches: map[string]gitversion.BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"main"},
			},
		},
		TagPrefix: "v",
	}

	latestVersion, _, err := gitversion.FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	expectedVersion, _ := semver.NewVersion("1.0.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}

func TestFindLatestVersion_FallbackWhenNoTagsOnSourceBranches(t *testing.T) {
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	_, err = r.CreateTag("v0.1.0", commit1, nil)
	assert.NoError(t, err)

	config := &gitversion.Config{
		Branches: map[string]gitversion.BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"develop"},
			},
		},
		TagPrefix: "v",
	}

	latestVersion, _, err := gitversion.FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	expectedVersion, _ := semver.NewVersion("0.1.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}

func TestFindLatestVersion_WithMultipleSourceBranches(t *testing.T) {
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Make a change before the second commit
	file2, err := fs.Create("CHANGELOG.md")
	assert.NoError(t, err)
	_, err = file2.Write([]byte("changelog content"))
	assert.NoError(t, err)
	err = file2.Close()
	assert.NoError(t, err)
	_, err = w.Add("CHANGELOG.md")
	assert.NoError(t, err)

	commit2, err := w.Commit("Second commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	_, err = r.CreateTag("v1.0.0", commit1, nil)
	assert.NoError(t, err)
	_, err = r.CreateTag("v2.0.0", commit2, nil)
	assert.NoError(t, err)

	config := &gitversion.Config{
		Branches: map[string]gitversion.BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"main", "develop"},
			},
		},
		TagPrefix: "v",
	}

	latestVersion, _, err := gitversion.FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	expectedVersion, _ := semver.NewVersion("2.0.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}
