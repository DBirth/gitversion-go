package cmd

// BranchConfig represents the configuration for a specific branch.
type BranchConfig struct {
	Mode string `yaml:"mode"`
	Tag  string `yaml:"tag"`
}

// Config represents the structure of the GitVersion.yml file.
type Config struct {
	NextVersion              string                `yaml:"next-version"`
	MajorVersionBumpMessage string                `yaml:"major-version-bump-message"`
	MinorVersionBumpMessage string                `yaml:"minor-version-bump-message"`
	PatchVersionBumpMessage string                `yaml:"patch-version-bump-message"`
	Branches                 map[string]BranchConfig `yaml:"branches"`
}
