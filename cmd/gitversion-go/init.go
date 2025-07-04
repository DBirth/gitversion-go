package main

import (
	"fmt"

	"gitversion-go/internal/fs"
	"gitversion-go/internal/gitversion"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var workflow string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a default GitVersion.yml file",
	RunE: func(_ *cobra.Command, _ []string) error {
		fs := fs.NewOsFs()
		return runInit(fs, workflow)
	},
}

func init() {
	initCmd.Flags().StringVar(&workflow, "workflow", "GitFlow", "Workflow template to use: GitFlow or GitHubFlow")
	rootCmd.AddCommand(initCmd)
}

func runInit(fs fs.Filesystem, workflow string) error {
	const configFileName = "GitVersion.yml"

	exists, err := fs.Exists(configFileName)
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

	err = fs.WriteFile(configFileName, []byte(template), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Successfully created a '%s' file for workflow '%s'.\n", configFileName, workflow)
	return nil
}
