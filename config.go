package main

// Config represents the structure of the GitVersion.yml file.
type Config struct {
	NextVersion              string `yaml:"next-version"`
	MajorVersionBumpMessage string `yaml:"major-version-bump-message"`
	MinorVersionBumpMessage string `yaml:"minor-version-bump-message"`
	PatchVersionBumpMessage string `yaml:"patch-version-bump-message"`
}
