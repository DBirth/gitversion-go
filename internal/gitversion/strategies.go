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
	Repository          *git.Repository
	Config              *Config
	CurrentBranchName   string
	BaseVersion         *semver.Version
	BaseVersionCommit   *object.Commit
	NextVersion         *semver.Version
	Bump                semverBump
	CommitsSinceLastTag int
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
	// Check for no-bump message first
	if config.NoBumpMessage != "" {
		if matched, _ := regexp.MatchString(config.NoBumpMessage, message); matched {
			return noBump
		}
	}

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
	err = commitIter.ForEach(func(c *object.Commit) error {
		if ctx.BaseVersionCommit != nil && c.Hash == ctx.BaseVersionCommit.Hash {
			return storer.ErrStop
		}
		commits = append(commits, c)
		return nil
	})
	if err != nil && err != storer.ErrStop {
		return false, err
	}
	ctx.CommitsSinceLastTag = len(commits)

	var highestBump = noBump
	for _, commit := range commits {
		bump := getBumpFromMessage(ctx.Config, commit.Message)
		if bump > highestBump {
			highestBump = bump
		}
	}

	branchConfig := ctx.Config.GetBranchConfig(ctx.CurrentBranchName)
	if highestBump == noBump && len(commits) > 0 && (branchConfig == nil || !branchConfig.PreventIncrement) {
		highestBump = patchBump
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
