package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// VersionVariables represents the version variables that can be used in the output.
type VersionVariables struct {
	Major      string `json:"Major"`
	Minor      string `json:"Minor"`
	Patch      string `json:"Patch"`
	Prerelease string `json:"Prerelease"`
	FullSemVer string `json:"FullSemVer"`
}

func buildVersionVariables(version semver.Version, branchName string, commitsSinceTag int, config *Config) VersionVariables {
	finalVersion := version
	var matchingBranchConfig *BranchConfig

	// Find matching branch configuration
	for pattern, bc := range config.Branches {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// In a real app, you might want to log this error
			continue
		}
		if re.MatchString(branchName) {
			matchingBranchConfig = &bc
			break
		}
	}

	if matchingBranchConfig != nil && commitsSinceTag > 0 {
		tag := matchingBranchConfig.Tag
		if tag == "use-branch-name" {
			// Sanitize branch name for use in prerelease tag
			sanitizedBranchName := strings.ReplaceAll(branchName, "/", "-")
			tag = sanitizedBranchName
		}

		if tag != "" {
			prerelease := fmt.Sprintf("%s.%d", tag, commitsSinceTag)
			v, err := finalVersion.SetPrerelease(prerelease)
			if err == nil {
				finalVersion = v
			}
		}
	}

	return VersionVariables{
		Major:      fmt.Sprintf("%d", finalVersion.Major()),
		Minor:      fmt.Sprintf("%d", finalVersion.Minor()),
		Patch:      fmt.Sprintf("%d", finalVersion.Patch()),
		Prerelease: finalVersion.Prerelease(),
		FullSemVer: finalVersion.String(),
	}
}
