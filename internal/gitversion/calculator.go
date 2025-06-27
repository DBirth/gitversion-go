// Package gitversion provides the core logic for calculating the next semantic version.
package gitversion

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// FindLatestVersion finds the latest semantic version tag in the repository.
func FindLatestVersion(r *git.Repository, config *Config) (*semver.Version, *object.Commit, error) {
	tagIter, err := r.Tags()
	if err != nil {
		return nil, nil, err
	}

	var latestVersion *semver.Version
	var latestTag *plumbing.Reference

	err = tagIter.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		prefixRegex, err := regexp.Compile(config.TagPrefix)
		if err != nil {
			// Handle invalid regex in config, maybe log it
			return nil
		}
		cleanedTagName := prefixRegex.ReplaceAllString(tagName, "")
		v, err := semver.NewVersion(cleanedTagName)
		if err != nil {
			return nil // Ignore non-semver tags
		}

		if latestVersion == nil || v.GreaterThan(latestVersion) {
			latestVersion = v
			latestTag = ref
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if latestTag == nil {
		return nil, nil, nil
	}

	// Get the commit for the latest tag
	latestTagCommit, err := r.CommitObject(latestTag.Hash())
	if err != nil {
		// This can happen with lightweight tags, try to resolve it
		tagObj, err := r.TagObject(latestTag.Hash())
		if err != nil {
			return nil, nil, err
		}
		latestTagCommit, err = r.CommitObject(tagObj.Target)
		if err != nil {
			return nil, nil, err
		}
	}

	return latestVersion, latestTagCommit, nil
}

type semverBump int

const (
	noBump semverBump = iota
	patchBump
	minorBump
	majorBump
)

// CalculateNextVersion calculates the next version based on the commit history.
func CalculateNextVersion(r *git.Repository, latestVersion *semver.Version, latestTagCommit *object.Commit, config *Config) (*semver.Version, int, error) {
	head, err := r.Head()
	if err != nil {
		return nil, 0, err
	}

	branchName := head.Name().Short()
	branchConfig := config.GetBranchConfig(branchName)

	// --- Count commits and determine bump ---
	commitIter, err := r.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, 0, err
	}

	var majorRegex, minorRegex, patchRegex, noBumpRegex *regexp.Regexp
	if config.MajorVersionBumpMessage != "" {
		majorRegex, _ = regexp.Compile(config.MajorVersionBumpMessage)
	}
	if config.MinorVersionBumpMessage != "" {
		minorRegex, _ = regexp.Compile(config.MinorVersionBumpMessage)
	}
	if config.PatchVersionBumpMessage != "" {
		patchRegex, _ = regexp.Compile(config.PatchVersionBumpMessage)
	}
	if config.NoBumpMessage != "" {
		noBumpRegex, _ = regexp.Compile(config.NoBumpMessage)
	}

	highestBump := noBump
	var commitsSinceTag int
	err = commitIter.ForEach(func(c *object.Commit) error {
		if latestTagCommit != nil && c.Hash == latestTagCommit.Hash {
			return storer.ErrStop
		}
		commitsSinceTag++

		// Check for no-bump message first
		if noBumpRegex != nil && noBumpRegex.MatchString(c.Message) {
			return nil
		}

		// For release branches, we don't bump based on commits.
		if branchConfig != nil && branchConfig.Mode == "semver-from-branch" {
			return nil
		}

		bump := noBump
		commitLines := strings.Split(c.Message, "\n")
		header := commitLines[0]
		conventionalCommitRegex := regexp.MustCompile(`^(feat|fix|build|chore|ci|docs|perf|refactor|revert|style|test)(\(.*\))?(!?):`)
		matches := conventionalCommitRegex.FindStringSubmatch(header)
		isBreaking := strings.Contains(c.Message, "BREAKING CHANGE:") || (len(matches) > 3 && matches[3] == "!")

		if isBreaking {
			bump = majorBump
		} else if len(matches) > 1 {
			switch matches[1] {
			case "feat":
				bump = minorBump
			case "fix":
				bump = patchBump
			}
		}

		if bump == noBump {
			if majorRegex != nil && majorRegex.MatchString(c.Message) {
				bump = majorBump
			} else if minorRegex != nil && minorRegex.MatchString(c.Message) {
				bump = minorBump
			} else if patchRegex != nil && patchRegex.MatchString(c.Message) {
				bump = patchBump
			}
		}
		highestBump = maxSemverBump(highestBump, bump)
		return nil
	})

	if err != nil && err != storer.ErrStop {
		return nil, 0, err
	}

	// --- Calculate next version ---

	// Handle semver-from-branch mode
	if branchConfig != nil && branchConfig.Mode == "semver-from-branch" {
		re := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
		matches := re.FindStringSubmatch(branchName)
		if len(matches) > 1 {
			baseVersion, err := semver.NewVersion(matches[1])
			if err == nil {
				nextVersion := *baseVersion
				// Always apply tag if configured, regardless of commits
				if branchConfig.Tag != "" {
					prerelease := branchConfig.Tag
					finalPrerelease := fmt.Sprintf("%s.%d", prerelease, commitsSinceTag)
					nextVersion, err = nextVersion.SetPrerelease(finalPrerelease)
					if err != nil {
						return nil, 0, err
					}
				}
				return &nextVersion, commitsSinceTag, nil
			}
		}
	}

	// Handle regular versioning
	bumpToApply := highestBump
	if commitsSinceTag > 0 && bumpToApply == noBump && (branchConfig == nil || branchConfig.Mode != "semver-from-branch") {
		bumpToApply = patchBump // Default to patch if there are commits but no bump messages
	}

	var nextVersion semver.Version
	if bumpToApply == majorBump {
		nextVersion = latestVersion.IncMajor()
	} else if bumpToApply == minorBump {
		nextVersion = latestVersion.IncMinor()
	} else if bumpToApply == patchBump {
		nextVersion = latestVersion.IncPatch()
	} else {
		nextVersion = *latestVersion
	}

	if branchConfig != nil && branchConfig.Tag != "" {
		prerelease := branchConfig.Tag
		if branchConfig.Tag == "use-branch-name" {
			prerelease = Sanitize(branchName)
		}
		var err error
		// If we have a pre-release tag and there was no version bump, we might need to append to a non-bumped version
		// But only if there are commits. If no commits, we just return the latest version.
		if commitsSinceTag > 0 {
			finalPrerelease := fmt.Sprintf("%s.%d", prerelease, commitsSinceTag)
			nextVersion, err = nextVersion.SetPrerelease(finalPrerelease)
			if err != nil {
				return nil, 0, err
			}
		} else if bumpToApply == noBump {
			// No commits, no bump, but a tag exists (e.g. on a release branch before any commits)
			// In this case, we often want version 0 of the pre-release
			finalPrerelease := fmt.Sprintf("%s.0", prerelease)
			nextVersion, err = nextVersion.SetPrerelease(finalPrerelease)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	return &nextVersion, commitsSinceTag, nil
}

// Sanitize prepares a branch name for use in a pre-release tag.
func Sanitize(branchName string) string {
	return strings.ReplaceAll(branchName, "/", "-")
}

// maxSemverBump returns the larger of two semverBumps.
func maxSemverBump(a, b semverBump) semverBump {
	if a > b {
		return a
	}
	return b
}
