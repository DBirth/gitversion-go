package gitversion

import "regexp"

// Config represents the structure of the GitVersion.yml file.
type Config struct {
	NextVersion             string                  `yaml:"next-version"`
	MajorVersionBumpMessage string                  `yaml:"major-version-bump-message"`
	MinorVersionBumpMessage string                  `yaml:"minor-version-bump-message"`
	PatchVersionBumpMessage string                  `yaml:"patch-version-bump-message"`
	NoBumpMessage           string                  `yaml:"no-bump-message"`
	TagPrefix               string                  `yaml:"tag-prefix"`
	Ignore                  []string                `yaml:"ignore,omitempty"`
	Increment               string                  `yaml:"increment,omitempty"`
	TagPreReleaseWeight     map[string]int          `yaml:"tag-pre-release-weight,omitempty"`
	Strategies              []string                `yaml:"strategies,omitempty"`
	Branches                map[string]BranchConfig `yaml:"branches"`
}

// BranchConfig represents the configuration for a specific branch.
type BranchConfig struct {
	Mode             string   `yaml:"mode"`
	Tag              string   `yaml:"tag"`
	Increment        string   `yaml:"increment,omitempty"`
	PreReleaseWeight int      `yaml:"pre-release-weight,omitempty"`
	SourceBranches   []string `yaml:"source-branches,omitempty"`
	Strategies       []string `yaml:"strategies,omitempty"`
	IsReleaseBranch  *bool    `yaml:"is-release-branch,omitempty"`
	PreventIncrement bool     `yaml:"prevent-increment,omitempty"`
}

// GetBranchConfig returns the configuration for a specific branch.
func (c *Config) GetBranchConfig(branchName string) *BranchConfig {
	if c.Branches == nil {
		return nil
	}

	var bestMatchConfig *BranchConfig
	var bestMatchPatternLength = -1

	for pattern, config := range c.Branches {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// Invalid regex in config, maybe log it
			continue
		}
		if re.MatchString(branchName) {
			// This is a match. Is it better than the previous best match?
			if len(pattern) > bestMatchPatternLength {
				bestMatchPatternLength = len(pattern)
				// Important: make a copy of config to avoid capturing loop variable
				branchConfigCopy := config
				bestMatchConfig = &branchConfigCopy
			}
		}
	}

	return bestMatchConfig
}
