// The main package for the gitversion-go command-line tool.
package main

import (
	"gitversion-go/internal/app"
	"gitversion-go/internal/fs"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var outputFormat string
var targetPath string

func init() {
	calculateCmd.Flags().StringVar(&outputFormat, "output", "default", "Output format (default, json)")
	calculateCmd.Flags().StringVar(&targetPath, "path", ".", "The path to the Git repository.")
	rootCmd.AddCommand(calculateCmd)
}

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculates the next version from the Git repository",
	Run: func(_ *cobra.Command, _ []string) {
		fileSystem := fs.NewOsFs()
		if err := app.RunCalculate(fileSystem, os.Stdout, targetPath, outputFormat); err != nil {
			log.Fatal(err)
		}
	},
}
