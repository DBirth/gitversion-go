package gitversion

import "strings"

// Config represents the structure of the GitVersion.yml file.
type Config struct {
	NextVersion              string `yaml:"next-version"`
	MajorVersionBumpMessage string `yaml:"major-version-bump-message"`
	MinorVersionBumpMessage string `yaml:"minor-version-bump-message"`
	PatchVersionBumpMessage string `yaml:"patch-version-bump-message"`
	Branches                 map[string]BranchConfig `yaml:"branches"`
}

// BranchConfig represents the configuration for a specific branch.
type BranchConfig struct {
	Mode            string `yaml:"mode"`
	Tag             string `yaml:"tag"`
	IsReleaseBranch *bool  `yaml:"is-release-branch,omitempty"`
}

// GetBranchConfig returns the configuration for a specific branch.
func (c *Config) GetBranchConfig(branchName string) *BranchConfig {
	if c.Branches == nil {
		return nil
	}

	// Direct match has priority
	if config, ok := c.Branches[branchName]; ok {
		return &config
	}

	// Wildcard match
	var bestMatchKey string
	var bestMatchLen int
	for key := range c.Branches {
		if strings.HasSuffix(key, "/*") {
			prefix := key[:len(key)-1] // e.g. "feature/"
			if strings.HasPrefix(branchName, prefix) {
				// Find the most specific match (longest prefix)
				if len(prefix) > bestMatchLen {
					bestMatchKey = key
					bestMatchLen = len(prefix)
				}
			}
		}
	}

	if bestMatchKey != "" {
		config := c.Branches[bestMatchKey]
		return &config
	}

	return nil
}
