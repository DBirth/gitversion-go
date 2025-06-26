package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitversion-go/pkg/fs"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)



var outputFormat string
var targetPath string

func init() {
	calculateCmd.Flags().StringVar(&outputFormat, "output", "default", "Output format (default, json)")
	calculateCmd.Flags().StringVar(&targetPath, "path", ".", "The path to the Git repository.")
	rootCmd.AddCommand(calculateCmd)
}

var gitPlainOpen = git.PlainOpen

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculates the next version from the Git repository",
	Run: func(cmd *cobra.Command, args []string) {
		fileSystem := fs.NewOsFs()
		if err := runCalculate(fileSystem, os.Stdout, targetPath, outputFormat); err != nil {
			log.Fatal(err)
		}
	},
}

func runCalculate(fs fs.Filesystem, out io.Writer, path, outputFormat string) error {
	// Load config
	var config Config
	configPath := filepath.Join(path, "GitVersion.yml")
	data, err := fs.ReadFile(configPath)
	if err == nil {
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return fmt.Errorf("failed to parse GitVersion.yml: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read GitVersion.yml: %w", err)
	}

	r, err := gitPlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %w", path, err)
	}

	head, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	latestVersion, latestTagCommit, err := findLatestVersion(r)
	if err != nil {
		return fmt.Errorf("failed to find latest version: %w", err)
	}

	if latestVersion == nil {
		fmt.Fprintln(out, "No semantic version tags found.")
		return nil
	}

	nextVersion, commitsSinceTag, err := calculateNextVersion(r, latestVersion, latestTagCommit, &config)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}

	branchName := head.Name().Short()
	vars := buildVersionVariables(*nextVersion, branchName, commitsSinceTag, &config)

	switch outputFormat {
	case "json":
		jsonOutput, err := json.Marshal(vars)
		if err != nil {
			return fmt.Errorf("failed to generate JSON output: %w", err)
		}
		fmt.Fprintln(out, string(jsonOutput))
	default:
		fmt.Fprintf(out, "Latest version found: %s\n", latestVersion.String())
		fmt.Fprintf(out, "Calculated next version: %s\n", vars.FullSemVer)
	}
	return nil
}

func findLatestVersion(r *git.Repository) (*semver.Version, *object.Commit, error) {
	tagIter, err := r.Tags()
	if err != nil {
		return nil, nil, err
	}

	var latestVersion *semver.Version
	var latestTag *plumbing.Reference

	err = tagIter.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		v, err := semver.NewVersion(tagName)
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

func calculateNextVersion(r *git.Repository, latestVersion *semver.Version, latestTagCommit *object.Commit, config *Config) (*semver.Version, int, error) {
	head, err := r.Head()
	if err != nil {
		return nil, 0, err
	}

	commitIter, err := r.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, 0, err
	}

	var majorRegex, minorRegex, patchRegex *regexp.Regexp
	if config.MajorVersionBumpMessage != "" {
		majorRegex, err = regexp.Compile(config.MajorVersionBumpMessage)
		if err != nil {
			log.Printf("Warning: Invalid regex for major-version-bump-message: %v. The pattern will be ignored.", err)
			majorRegex = nil
		}
	}
	if config.MinorVersionBumpMessage != "" {
		minorRegex, err = regexp.Compile(config.MinorVersionBumpMessage)
		if err != nil {
			log.Printf("Warning: Invalid regex for minor-version-bump-message: %v. The pattern will be ignored.", err)
			minorRegex = nil
		}
	}
	if config.PatchVersionBumpMessage != "" {
		patchRegex, err = regexp.Compile(config.PatchVersionBumpMessage)
		if err != nil {
			log.Printf("Warning: Invalid regex for patch-version-bump-message: %v. The pattern will be ignored.", err)
			patchRegex = nil
		}
	}



	highestBump := noBump
	var commitsSinceTag int
	err = commitIter.ForEach(func(c *object.Commit) error {
		if latestTagCommit != nil && c.Hash == latestTagCommit.Hash {
			return storer.ErrStop
		}
		commitsSinceTag++

		// Determine bump from legacy messages
		if majorRegex != nil && majorRegex.MatchString(c.Message) {
			highestBump = max(highestBump, majorBump)
		}
		if minorRegex != nil && minorRegex.MatchString(c.Message) {
			highestBump = max(highestBump, minorBump)
		}
		if patchRegex != nil && patchRegex.MatchString(c.Message) {
			highestBump = max(highestBump, patchBump)
		}

		// Determine bump from conventional commits
		commitLines := strings.Split(c.Message, "\n")
		header := commitLines[0]

		if strings.Contains(c.Message, "BREAKING CHANGE:") || strings.Contains(header, "!:") {
			highestBump = max(highestBump, majorBump)
		}
		if strings.HasPrefix(header, "feat") {
			highestBump = max(highestBump, minorBump)
		}
		if strings.HasPrefix(header, "fix") {
			highestBump = max(highestBump, patchBump)
		}

		return nil
	})

	if err != nil && err != storer.ErrStop {
		return nil, 0, err
	}

	nextVersion := *latestVersion
	switch highestBump {
	case majorBump:
		nextVersion = latestVersion.IncMajor()
	case minorBump:
		nextVersion = latestVersion.IncMinor()
	case patchBump:
		nextVersion = latestVersion.IncPatch()
	default:
		// If there are commits since the last tag, but not one that bumps major/minor/patch, we should bump patch
		if commitsSinceTag > 0 {
			nextVersion = latestVersion.IncPatch()
		}
	}

	return &nextVersion, commitsSinceTag, nil
}

func max(a, b semverBump) semverBump {
	if a > b {
		return a
	}
	return b
}

