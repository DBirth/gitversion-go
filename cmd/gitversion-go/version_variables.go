package main

import (
	"fmt"
	"strings"

	"gitversion-go/internal/gitversion"

	"github.com/Masterminds/semver/v3"
)

// VersionVariables represents the version variables that can be used in the output.
type VersionVariables struct {
	Major         string `json:"Major"`
	Minor         string `json:"Minor"`
	Patch         string `json:"Patch"`
	PreReleaseTag string `json:"PreReleaseTag"`
	FullSemVer    string `json:"FullSemVer"`
}

func buildVersionVariables(version semver.Version, branchName string, commitsSinceTag int, config *gitversion.Config) VersionVariables {
	finalVersion := version
	var matchingBranchConfig *gitversion.BranchConfig

	matchingBranchConfig = config.GetBranchConfig(branchName)

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
		Major:         fmt.Sprintf("%d", finalVersion.Major()),
		Minor:         fmt.Sprintf("%d", finalVersion.Minor()),
		Patch:         fmt.Sprintf("%d", finalVersion.Patch()),
		PreReleaseTag: finalVersion.Prerelease(),
		FullSemVer:    finalVersion.String(),
	}
}
