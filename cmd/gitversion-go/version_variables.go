package main

// VersionVariables represents the version variables that can be used in the output.
type VersionVariables struct {
	Major         string `json:"Major"`
	Minor         string `json:"Minor"`
	Patch         string `json:"Patch"`
	PreReleaseTag string `json:"PreReleaseTag"`
	FullSemVer    string `json:"FullSemVer"`
}
