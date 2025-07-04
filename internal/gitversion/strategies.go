package gitversion

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// VersionContext holds all the information needed for versioning strategies.
type VersionContext struct {
	Repository            *git.Repository
	Config                *Config
	CurrentBranchName     string
	BaseVersion           *semver.Version
	BaseVersionCommit     *object.Commit
	NextVersion           *semver.Version
	Bump                  semverBump
	CommitsSinceLastTag   int
	FormattedCommitDates  []string // commit dates formatted per config.CommitDateFormat
	MergeCommitIndices    []int    // indices in commits slice that match merge-message-formats
}

type semverBump int

const (
	noBump semverBump = iota
	patchBump
	minorBump
	majorBump
)

// VersioningStrategy defines the interface for a versioning strategy.
type VersioningStrategy interface {
	Execute(ctx *VersionContext) (bool, error)
}

// StrategyExecutor manages the execution of a list of versioning strategies.
type StrategyExecutor struct {
	Strategies []VersioningStrategy
}

// NewStrategyExecutor creates a new StrategyExecutor with the given strategies.
func NewStrategyExecutor(strategies []VersioningStrategy) *StrategyExecutor {
	return &StrategyExecutor{Strategies: strategies}
}

// ExecuteStrategies runs the strategies in order until one of them succeeds.
func (e *StrategyExecutor) ExecuteStrategies(ctx *VersionContext) error {
	for _, strategy := range e.Strategies {
		success, err := strategy.Execute(ctx)
		if err != nil {
			return err
		}
		if success {
			break
		}
	}
	return nil
}

// FindLatestTagStrategy finds the latest semantic version tag in the repository.
type FindLatestTagStrategy struct{}

func (s *FindLatestTagStrategy) Execute(ctx *VersionContext) (bool, error) {
	if ctx.BaseVersion != nil {
		return false, nil
	}

	latestVersion, latestTagCommit, err := FindLatestVersion(ctx.Repository, ctx.Config, ctx.CurrentBranchName)
	if err != nil {
		return false, err
	}

	if latestVersion != nil {
		ctx.BaseVersion = latestVersion
		ctx.BaseVersionCommit = latestTagCommit
	}

	return false, nil
}

// getBumpFromMessage analyzes a commit message and returns the bump type.
func getBumpFromMessage(config *Config, message string) semverBump {
	// Conventional commits
	commitLines := strings.Split(message, "\n")
	header := commitLines[0]
	conventionalCommitRegex := regexp.MustCompile(`^(feat|fix|build|chore|ci|docs|perf|refactor|revert|style|test)(\(.*\))?(!?):`)
	matches := conventionalCommitRegex.FindStringSubmatch(header)
	isBreaking := strings.Contains(message, "BREAKING CHANGE:") || (len(matches) > 3 && matches[3] == "!")

	if isBreaking {
		return majorBump
	}
	if len(matches) > 1 {
		switch matches[1] {
		case "feat":
			return minorBump
		case "fix":
			return patchBump
		}
	}

	// Custom regexes
	if config.MajorVersionBumpMessage != "" {
		if matched, _ := regexp.MatchString(config.MajorVersionBumpMessage, message); matched {
			return majorBump
		}
	}
	if config.MinorVersionBumpMessage != "" {
		if matched, _ := regexp.MatchString(config.MinorVersionBumpMessage, message); matched {
			return minorBump
		}
	}
	if config.PatchVersionBumpMessage != "" {
		if matched, _ := regexp.MatchString(config.PatchVersionBumpMessage, message); matched {
			return patchBump
		}
	}

	return noBump
}

// IncrementFromCommitsStrategy increments the base version based on commit messages.
type IncrementFromCommitsStrategy struct{}

func (s *IncrementFromCommitsStrategy) Execute(ctx *VersionContext) (bool, error) {
	// Determine commit date format
	commitDateFormat := ctx.Config.CommitDateFormat
	if commitDateFormat == "" {
		commitDateFormat = "2006-01-02T15:04:05Z07:00" // ISO8601 default
	}
	// Prepare merge commit regexes
	var mergeRegexes []*regexp.Regexp
	if len(ctx.Config.MergeMessageFormats) > 0 {
		for _, pat := range ctx.Config.MergeMessageFormats {
			re, err := regexp.Compile(pat)
			if err == nil {
				mergeRegexes = append(mergeRegexes, re)
			}
		}
	} else {
		// Default patterns: GitHub/GitLab/Bitbucket
		defaultPatterns := []string{
			`^Merge pull request #`,
			`^Merge branch '`,
			`^Merged in `,
		}
		for _, pat := range defaultPatterns {
			re, _ := regexp.Compile(pat)
			mergeRegexes = append(mergeRegexes, re)
		}
	}

	if ctx.BaseVersion == nil || ctx.NextVersion != nil {
		return false, nil
	}

	head, err := ctx.Repository.Head()
	if err != nil {
		return false, err
	}

	commitIter, err := ctx.Repository.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return false, err
	}
	defer commitIter.Close()

	var commits []*object.Commit
	ctx.FormattedCommitDates = nil
	ctx.MergeCommitIndices = nil
	err = commitIter.ForEach(func(c *object.Commit) error {
		if ctx.BaseVersionCommit != nil && c.Hash == ctx.BaseVersionCommit.Hash {
			return storer.ErrStop
		}
		// Ignore commit if SHA is in config.Ignore
		ignored := false
		for _, sha := range ctx.Config.Ignore {
			if c.Hash.String() == sha {
				ignored = true
				break
			}
		}
		if ignored {
			return nil
		}
		commits = append(commits, c)
		// Format and store commit date
		ctx.FormattedCommitDates = append(ctx.FormattedCommitDates, c.Committer.When.Format(commitDateFormat))
		// Check for merge commit
		for _, re := range mergeRegexes {
			if re.MatchString(c.Message) {
				ctx.MergeCommitIndices = append(ctx.MergeCommitIndices, len(commits)-1)
				break
			}
		}
		return nil
	})
	if err != nil && err != storer.ErrStop {
		return false, err
	}
	// Reverse commits so that commits[0] is the most recent
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	ctx.CommitsSinceLastTag = len(commits)

	// If the most recent commit matches no-bump-message, do not bump at all
	if len(commits) > 0 {
		if strings.Contains(commits[0].Message, "+semver: none") || strings.Contains(commits[0].Message, "+semver: skip") {
			ctx.Bump = noBump
			ctx.NextVersion = ctx.BaseVersion
			return true, nil // No bump if no-bump-message found
		}
	}
	if len(commits) > 0 && ctx.Config.NoBumpMessage != "" {
		matched, _ := regexp.MatchString(ctx.Config.NoBumpMessage, commits[0].Message)
		if matched {
			ctx.Bump = noBump
			ctx.NextVersion = ctx.BaseVersion
			return true, nil // No bump if no-bump-message found
		}
	}
	var highestBump = noBump
	for _, commit := range commits {
		bump := getBumpFromMessage(ctx.Config, commit.Message)
		if bump > highestBump {
			highestBump = bump
		}
	}
	branchConfig := ctx.Config.GetBranchConfig(ctx.CurrentBranchName);
	// Handle semver-from-branch mode (for release branches)
	if branchConfig != nil && branchConfig.Mode == "semver-from-branch" {
		// Try to parse version from branch name (e.g., release/1.2.3)
		re := regexp.MustCompile(`\d+\.\d+\.\d+`)
		match := re.FindString(ctx.CurrentBranchName)
		if match != "" {
			v, err := semver.NewVersion(match)
			if err == nil {
				// Add pre-release tag if specified
				ver := *v
				if branchConfig.Tag != "" {
					ver, _ = ver.SetPrerelease(branchConfig.Tag + ".1")
				}
				ctx.NextVersion = &ver
				ctx.Bump = noBump
				return true, nil
			}
		}
	}
	// Use increment setting if no bump detected
	if highestBump == noBump && len(commits) > 0 && (branchConfig == nil || !branchConfig.PreventIncrement) {
		// Only apply increment setting for the *first* commit after the tag
		increment := ""
		if branchConfig != nil && branchConfig.Increment != "" {
			increment = strings.ToLower(branchConfig.Increment)
		} else if ctx.Config.Increment != "" {
			increment = strings.ToLower(ctx.Config.Increment)
		}
		if increment != "" {
			switch increment {
			case "major":
				highestBump = majorBump
			case "minor":
				highestBump = minorBump
			case "patch":
				highestBump = patchBump
			default:
				highestBump = patchBump // Default to patch if not specified
			}
		} else {
			highestBump = patchBump
		}
	}
	ctx.Bump = highestBump
	if highestBump != noBump {
		nextVersion := *ctx.BaseVersion
		switch highestBump {
		case majorBump:
			nextVersion = nextVersion.IncMajor()
		case minorBump:
			nextVersion = nextVersion.IncMinor()
		case patchBump:
			nextVersion = nextVersion.IncPatch()
		}
		ctx.NextVersion = &nextVersion
		return true, nil // Strategy produced a version
	}
	return false, nil // No increment found
}

// ConfiguredNextVersionStrategy provides a version based on the 'next-version' configuration.
type ConfiguredNextVersionStrategy struct{}

func (s *ConfiguredNextVersionStrategy) Execute(ctx *VersionContext) (bool, error) {
	if ctx.BaseVersion != nil || ctx.NextVersion != nil {
		return false, nil
	}

	if ctx.Config.NextVersion == "" {
		return false, nil
	}

	v, err := semver.NewVersion(ctx.Config.NextVersion)
	if err != nil {
		return false, fmt.Errorf("invalid next-version: %w", err)
	}

	ctx.NextVersion = v
	return true, nil
}

var strategyFactories = map[string]func() VersioningStrategy{
	"find-latest-tag":          func() VersioningStrategy { return &FindLatestTagStrategy{} },
	"increment-from-commits":   func() VersioningStrategy { return &IncrementFromCommitsStrategy{} },
	"configured-next-version":  func() VersioningStrategy { return &ConfiguredNextVersionStrategy{} },
}

func BuildStrategies(config *Config, branchName string) ([]VersioningStrategy, error) {
	branchConfig := config.GetBranchConfig(branchName)

	strategyNames := config.Strategies
	if branchConfig != nil && len(branchConfig.Strategies) > 0 {
		strategyNames = branchConfig.Strategies
	}

	if len(strategyNames) == 0 {
		strategyNames = []string{"find-latest-tag", "increment-from-commits", "configured-next-version"}
	}

	var strategies []VersioningStrategy
	for _, name := range strategyNames {
		factory, ok := strategyFactories[name]
		if !ok {
			return nil, fmt.Errorf("unknown strategy: %s", name)
		}
		strategies = append(strategies, factory())
	}
	return strategies, nil
}
