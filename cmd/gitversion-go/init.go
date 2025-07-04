package main

import (
	"gitversion-go/internal/app"
	"gitversion-go/internal/fs"

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
		return app.RunInit(fs, workflow)
	},
}

func init() {
	initCmd.Flags().StringVar(&workflow, "workflow", "GitFlow", "Workflow template to use: GitFlow or GitHubFlow")
	rootCmd.AddCommand(initCmd)
}
