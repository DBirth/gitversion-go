package gitversion

import (
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestFindLatestVersionWithSourceBranches(t *testing.T) {
	// Setup in-memory git repository
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create a file to have a non-empty commit
	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	// Create the first commit
	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Create main branch pointing to the first commit
	mainRefName := plumbing.NewBranchReferenceName("main")
	mainRef := plumbing.NewHashReference(mainRefName, commit1)
	err = r.Storer.SetReference(mainRef)
	assert.NoError(t, err)

	// Add a version tag to the commit on main
	_, err = r.CreateTag("v1.0.0", commit1, nil)
	assert.NoError(t, err)

	// Create and checkout a feature branch
	featureRefName := plumbing.NewBranchReferenceName("feature/new-stuff")
	err = w.Checkout(&git.CheckoutOptions{Branch: featureRefName, Create: true})
	assert.NoError(t, err)

	// Create a new commit on the feature branch
	file, err = fs.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	_, err = file.Write([]byte("\nfeature content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	_, err = w.Commit("Feature commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Configure source-branches for the feature branch
	config := &Config{
		Branches: map[string]BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"main"},
			},
		},
		TagPrefix: "v",
	}

	// Find the latest version for the feature branch
	latestVersion, _, err := FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	// Assert that the version from the source branch is found
	expectedVersion, _ := semver.NewVersion("1.0.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}

func TestFindLatestVersion_FallbackWhenNoTagsOnSourceBranches(t *testing.T) {
	// Setup in-memory git repository
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create a file to have a non-empty commit
	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	// Create the first commit and a fallback tag
	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)
	_, err = r.CreateTag("v0.1.0", commit1, nil)
	assert.NoError(t, err)

	// Create develop branch pointing to the first commit
	developRefName := plumbing.NewBranchReferenceName("develop")
	developRef := plumbing.NewHashReference(developRefName, commit1)
	err = r.Storer.SetReference(developRef)
	assert.NoError(t, err)

	// Create and checkout a feature branch from develop
	featureRefName := plumbing.NewBranchReferenceName("feature/new-stuff")
	err = w.Checkout(&git.CheckoutOptions{Branch: featureRefName, Create: true, Hash: commit1})
	assert.NoError(t, err)

	// Create a new commit on the feature branch
	file, err = fs.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	_, err = file.Write([]byte("\nfeature content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	_, err = w.Commit("Feature commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Configure source-branches for the feature branch, but develop has no tags
	config := &Config{
		Branches: map[string]BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"develop"},
			},
		},
		TagPrefix: "v",
	}

	// Find the latest version for the feature branch
	latestVersion, _, err := FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	// Assert that the version from the global search is found
	expectedVersion, _ := semver.NewVersion("0.1.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}

func TestFindLatestVersion_WithMultipleSourceBranches(t *testing.T) {
	// Setup in-memory git repository
	storer := memory.NewStorage()
	fs := memfs.New()
	r, err := git.Init(storer, fs)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create a file to have a non-empty commit
	file, err := fs.Create("README.md")
	assert.NoError(t, err)
	_, err = file.Write([]byte("initial content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)

	// Create the first commit for main
	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Create main branch and tag it
	mainRefName := plumbing.NewBranchReferenceName("main")
	mainRef := plumbing.NewHashReference(mainRefName, commit1)
	err = r.Storer.SetReference(mainRef)
	assert.NoError(t, err)
	_, err = r.CreateTag("v1.0.0", commit1, nil)
	assert.NoError(t, err)

	// Create a second commit for develop
	file, err = fs.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	_, err = file.Write([]byte("\ndevelop content"))
	assert.NoError(t, err)
	err = file.Close()
	assert.NoError(t, err)
	_, err = w.Add("README.md")
	assert.NoError(t, err)
	commit2, err := w.Commit("Develop commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	assert.NoError(t, err)

	// Create develop branch and tag it with a higher version
	developRefName := plumbing.NewBranchReferenceName("develop")
	developRef := plumbing.NewHashReference(developRefName, commit2)
	err = r.Storer.SetReference(developRef)
	assert.NoError(t, err)
	_, err = r.CreateTag("v1.1.0", commit2, nil)
	assert.NoError(t, err)

	// Create a feature branch
	featureRefName := plumbing.NewBranchReferenceName("feature/new-stuff")
	err = w.Checkout(&git.CheckoutOptions{Branch: featureRefName, Create: true, Hash: commit2})
	assert.NoError(t, err)

	// Configure source-branches for the feature branch
	config := &Config{
		Branches: map[string]BranchConfig{
			"feature/new-stuff": {
				SourceBranches: []string{"main", "develop"},
			},
		},
		TagPrefix: "v",
	}

	// Find the latest version for the feature branch
	latestVersion, _, err := FindLatestVersion(r, config, "feature/new-stuff")
	assert.NoError(t, err)

	// Assert that the highest version from the source branches is found
	expectedVersion, _ := semver.NewVersion("1.1.0")
	assert.NotNil(t, latestVersion)
	assert.True(t, latestVersion.Equal(expectedVersion), "Expected version %s, but got %s", expectedVersion, latestVersion)
}
