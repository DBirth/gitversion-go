package cmd

import (
	"fmt"
	"gitversion-go/pkg/fs"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a default GitVersion.yml file",
	RunE: func(cmd *cobra.Command, args []string) error {
		fs := fs.NewOsFs()
		return runInit(fs)
	},
}

func runInit(fs fs.Filesystem) error {
	const configFileName = "GitVersion.yml"

	exists, err := fs.Exists(configFileName)
	if err != nil {
		return fmt.Errorf("failed to check for existing config file: %w", err)
	}

	if exists {
		fmt.Printf("A '%s' file already exists.\n", configFileName)
		return nil
	}

	defaultConfig := `next-version: 0.1.0
major-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(breaking|major))"
minor-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(feature|minor))"
patch-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(fix|patch))"
branches:
  main:
    tag: ""
  develop:
    mode: ContinuousDeployment
    tag: alpha
  release:
    mode: semver-from-branch
    tag: beta
  feature:
    tag: use-branch-name
  hotfix:
    mode: semver-from-branch
    tag: beta
`

	err = fs.WriteFile(configFileName, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Successfully created '%s'.\n", configFileName)
	return nil
}
