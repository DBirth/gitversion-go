// Package app provides core logic for CLI and tests.
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitversion-go/internal/fs"
	"gitversion-go/internal/gitversion"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"
)

// RunInit creates a GitVersion.yml config file for the given workflow.
func RunInit(fsys fs.Filesystem, workflow string) error {
	const configFileName = "GitVersion.yml"

	exists, err := fsys.Exists(configFileName)
	if err != nil {
		return fmt.Errorf("failed to check for existing config file: %w", err)
	}

	if exists {
		fmt.Printf("A '%s' file already exists.\n", configFileName)
		return nil
	}

	template := gitversion.GetWorkflowTemplate(workflow)
	if template == "" {
		return fmt.Errorf("unknown workflow: %s", workflow)
	}

	err = fsys.WriteFile(configFileName, []byte(template), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Successfully created a '%s' file for workflow '%s'.\n", configFileName, workflow)
	return nil
}

// RunCalculate calculates the next version and writes output to the writer.
func RunCalculate(fsys fs.Filesystem, out io.Writer, path, outputFormat string) error {
	var config gitversion.Config
	configPath := filepath.Join(path, "GitVersion.yml")

	data, err := fsys.ReadFile(configPath)
	if err == nil {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("failed to parse GitVersion.yml: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read GitVersion.yml: %w", err)
	}

	r, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %w", path, err)
	}

	head, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()
	nextVersion, commitsSinceTag, err := gitversion.CalculateNextVersion(r, &config, branchName)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}

	vars := buildVersionVariables(nextVersion, branchName, commitsSinceTag, &config)

	switch outputFormat {
	case "json":
		jsonOutput, err := json.Marshal(vars)
		if err != nil {
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}
		if _, err := fmt.Fprintln(out, string(jsonOutput)); err != nil {
			return err
		}
	default:
		if _, err := fmt.Fprintf(out, "Calculated next version: %s\n", vars.FullSemVer); err != nil {
			return err
		}
	}
	return nil
}

// VersionVariables holds version information for output formats.
type VersionVariables struct {
	Major         string `json:"Major"`
	Minor         string `json:"Minor"`
	Patch         string `json:"Patch"`
	PreReleaseTag string `json:"PreReleaseTag"`
	FullSemVer    string `json:"FullSemVer"`
}

func buildVersionVariables(version *semver.Version, branchName string, commitsSinceTag int, config *gitversion.Config) VersionVariables {
	finalVersion := *version
	matchingBranchConfig := config.GetBranchConfig(branchName)

	if matchingBranchConfig != nil && commitsSinceTag > 0 {
		tag := matchingBranchConfig.Tag
		if tag == "use-branch-name" {
			// Sanitize branch name for use in prerelease tag
			sanitizedBranchName := branchName
			tag = sanitizedBranchName
		}

		if tag != "" {
			// Sanitize branch name for prerelease: replace slashes with dashes
			sanitizedTag := strings.ReplaceAll(tag, "/", "-")
			var prerelease string
			if matchingBranchConfig.PreReleaseWeight > 0 {
				prerelease = fmt.Sprintf("%s.%d.%d", sanitizedTag, matchingBranchConfig.PreReleaseWeight, commitsSinceTag)
			} else {
				prerelease = fmt.Sprintf("%s.%d", sanitizedTag, commitsSinceTag)
			}
			v, err := semver.NewVersion(fmt.Sprintf("%d.%d.%d-%s", finalVersion.Major(), finalVersion.Minor(), finalVersion.Patch(), prerelease))
			if err == nil {
				finalVersion = *v
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
