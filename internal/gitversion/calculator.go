// Package gitversion provides the core logic for calculating the next semantic version.
package gitversion

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CalculateNextVersion calculates the next version based on the commit history using a strategy-based approach.
func CalculateNextVersion(r *git.Repository, config *Config, currentBranchName string) (*semver.Version, int, error) {
	strategies, err := BuildStrategies(config, currentBranchName)
	if err != nil {
		return nil, 0, err
	}

	executor := NewStrategyExecutor(strategies)
	ctx := &VersionContext{
		Repository:        r,
		Config:            config,
		CurrentBranchName: currentBranchName,
	}

	if err := executor.ExecuteStrategies(ctx); err != nil {
		return nil, 0, err
	}

	if ctx.NextVersion == nil {
		if ctx.BaseVersion != nil {
			return ctx.BaseVersion, ctx.CommitsSinceLastTag, nil
		}
		// Fallback to 0.1.0 if no version could be determined.
		v := semver.MustParse("0.1.0")
		return v, 0, nil
	}

	return ctx.NextVersion, ctx.CommitsSinceLastTag, nil
}

// FindLatestVersion finds the latest semantic version tag in the repository.
// It first checks the source branches of the current branch, if any.
// If no version is found on the source branches, it searches all tags.
func FindLatestVersion(r *git.Repository, config *Config, currentBranchName string) (*semver.Version, *object.Commit, error) {
	branchConfig := config.GetBranchConfig(currentBranchName)
	if branchConfig != nil && len(branchConfig.SourceBranches) > 0 {
		latestVersion, latestTagCommit, err := findVersionOnBranches(r, config, branchConfig.SourceBranches)
		if err != nil {
			return nil, nil, err
		}
		if latestVersion != nil {
			return latestVersion, latestTagCommit, nil
		}
	}

	return findLatestVersionAllTags(r, config)
}

func findVersionOnBranches(r *git.Repository, config *Config, branchNames []string) (*semver.Version, *object.Commit, error) {
	var versions []*semver.Version
	tagCommitMap := make(map[*semver.Version]*object.Commit)

	for _, branchName := range branchNames {
		branchRef, err := r.Reference(plumbing.NewBranchReferenceName(branchName), true)
		if err != nil {
			if errors.Is(err, plumbing.ErrReferenceNotFound) {
				continue
			}
			return nil, nil, err
		}

		commit, err := r.CommitObject(branchRef.Hash())
		if err != nil {
			return nil, nil, err
		}

		tags, err := getTags(r)
		if err != nil {
			return nil, nil, err
		}

		commitIter, err := r.Log(&git.LogOptions{From: commit.Hash})
		if err != nil {
			return nil, nil, err
		}

		commitTags := make(map[plumbing.Hash][]string)
		for _, tag := range tags {
			// TODO: This is inefficient. We should get the commit from the tag ref instead.
			commitTags[tag.Hash()] = append(commitTags[tag.Hash()], tag.Name().Short())
		}

		err = commitIter.ForEach(func(c *object.Commit) error {
			if tagNames, ok := commitTags[c.Hash]; ok {
				for _, tagName := range tagNames {
					v, err := semver.NewVersion(tagName)
					if err == nil {
						versions = append(versions, v)
						tagCommitMap[v] = c
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}

	if len(versions) == 0 {
		return nil, nil, nil
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].LessThan(versions[j])
	})

	latestVersion := versions[len(versions)-1]
	return latestVersion, tagCommitMap[latestVersion], nil
}

func findLatestVersionAllTags(r *git.Repository, config *Config) (*semver.Version, *object.Commit, error) {
	tagRefs, err := r.Tags()
	if err != nil {
		return nil, nil, err
	}

	var versions []*semver.Version
	tagCommitMap := make(map[*semver.Version]*object.Commit)

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		v, err := semver.NewVersion(ref.Name().Short())
		if err == nil {
			commit, err := getCommitFromTag(r, ref)
			if err != nil {
				// Cannot resolve tag, skip
				return nil
			}
			versions = append(versions, v)
			tagCommitMap[v] = commit
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if len(versions) == 0 {
		return nil, nil, nil
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].LessThan(versions[j])
	})

	latestVersion := versions[len(versions)-1]
	return latestVersion, tagCommitMap[latestVersion], nil
}

func getCommitFromTag(r *git.Repository, ref *plumbing.Reference) (*object.Commit, error) {
	obj, err := r.Object(plumbing.AnyObject, ref.Hash())
	if err != nil {
		return nil, err
	}

	switch o := obj.(type) {
	case *object.Commit:
		return o, nil
	case *object.Tag:
		return o.Commit()
	default:
		return nil, fmt.Errorf("unexpected object type %s for tag %s", o.Type(), ref.Name())
	}
}

func getTags(r *git.Repository) ([]*plumbing.Reference, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}

	var tags []*plumbing.Reference
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

