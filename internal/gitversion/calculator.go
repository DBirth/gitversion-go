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

	isGreaterThan := func(v1, v2 *semver.Version) bool {
		// If major, minor, or patch are different, standard comparison is enough.
		if v1.Major() != v2.Major() || v1.Minor() != v2.Minor() || v1.Patch() != v2.Patch() {
			return v1.GreaterThan(v2)
		}

		// If we are here, major.minor.patch are the same. Compare pre-release tags.
		prerelease1 := v1.Prerelease()
		prerelease2 := v2.Prerelease()

		// A version without a prerelease tag has higher precedence
		if prerelease1 == "" && prerelease2 != "" {
			return true
		}
		if prerelease1 != "" && prerelease2 == "" {
			return false
		}
		if prerelease1 == "" && prerelease2 == "" {
			return false // they are equal
		}

		// Extract the tag name (e.g., "alpha" from "alpha.1")
		tag1 := strings.Split(prerelease1, ".")[0]
		tag2 := strings.Split(prerelease2, ".")[0]

		weight1, ok1 := config.TagPreReleaseWeight[tag1]
		weight2, ok2 := config.TagPreReleaseWeight[tag2]

		// If both tags have weights defined, compare them.
		if ok1 && ok2 {
			if weight1 != weight2 {
				return weight1 > weight2
			}
		}

		// Fallback to standard semver comparison
		return v1.GreaterThan(v2)
	}

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

		if latestVersion == nil || isGreaterThan(v, latestVersion) {
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

	// Create a set for ignored SHAs for efficient lookup
	ignoredShaSet := make(map[string]struct{})
	if config.Ignore != nil {
		for _, sha := range config.Ignore {
			ignoredShaSet[sha] = struct{}{}
		}
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

	// Determine the effective increment strategy
	effectiveIncrement := config.Increment
	if branchConfig != nil && branchConfig.Increment != "" {
		effectiveIncrement = branchConfig.Increment
	}

	err = commitIter.ForEach(func(c *object.Commit) error {
		if latestTagCommit != nil && c.Hash == latestTagCommit.Hash {
			return storer.ErrStop
		}

		// Check if the commit SHA is in the ignore list
		if _, ok := ignoredShaSet[c.Hash.String()]; ok {
			return nil
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

	// Apply branch-level increment strategy if not inheriting
	if effectiveIncrement != "" && effectiveIncrement != "Inherit" && commitsSinceTag > 0 {
		branchBump := noBump
		switch effectiveIncrement {
		case "Major":
			branchBump = majorBump
		case "Minor":
			branchBump = minorBump
		case "Patch":
			branchBump = patchBump
		case "None":
			// Do nothing
		}
		highestBump = maxSemverBump(highestBump, branchBump)
	}

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
					var finalPrerelease string
					if branchConfig.PreReleaseWeight > 0 {
						finalPrerelease = fmt.Sprintf("%s.%d.%d", prerelease, branchConfig.PreReleaseWeight, commitsSinceTag)
					} else {
						finalPrerelease = fmt.Sprintf("%s.%d", prerelease, commitsSinceTag)
					}
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
		if commitsSinceTag > 0 {
			var finalPrerelease string
			if branchConfig.PreReleaseWeight > 0 {
				finalPrerelease = fmt.Sprintf("%s.%d.%d", prerelease, branchConfig.PreReleaseWeight, commitsSinceTag)
			} else {
				finalPrerelease = fmt.Sprintf("%s.%d", prerelease, commitsSinceTag)
			}
			nextVersion, err = nextVersion.SetPrerelease(finalPrerelease)
			if err != nil {
				return nil, 0, err
			}
		} else if nextVersion.Prerelease() == "" {
			var finalPrerelease string
			if branchConfig.PreReleaseWeight > 0 {
				finalPrerelease = fmt.Sprintf("%s.%d.%d", prerelease, branchConfig.PreReleaseWeight, 0)
			} else {
				finalPrerelease = fmt.Sprintf("%s.%d", prerelease, 0)
			}
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
