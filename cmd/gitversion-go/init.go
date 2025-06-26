package main

import (
	"fmt"

	"gitversion-go/internal/gitversion"
	"gitversion-go/internal/pkg/fs"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a default GitVersion.yml file",
	RunE: func(_ *cobra.Command, _ []string) error {
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

	err = fs.WriteFile(configFileName, []byte(gitversion.DefaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	fmt.Printf("Successfully created a '%s' file.\n", configFileName)
	return nil
}
