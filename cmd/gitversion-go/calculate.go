// The main package for the gitversion-go command-line tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"gitversion-go/internal/fs"
	"gitversion-go/internal/gitversion"
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
	Run: func(_ *cobra.Command, _ []string) {
		fileSystem := fs.NewOsFs()
		if err := runCalculate(fileSystem, os.Stdout, targetPath, outputFormat); err != nil {
			log.Fatal(err)
		}
	},
}

func runCalculate(fs fs.Filesystem, out io.Writer, path, outputFormat string) error {
	// Load config
	var config gitversion.Config
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
	branchName := head.Name().Short()

	nextVersion, commitsSinceTag, err := gitversion.CalculateNextVersion(r, &config, branchName)
	if err != nil {
		return fmt.Errorf("failed to calculate next version: %w", err)
	}

	vars := buildVersionVariables(*nextVersion, branchName, commitsSinceTag, &config)

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
