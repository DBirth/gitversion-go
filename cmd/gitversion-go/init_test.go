package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInit_FileDoesNotExist(t *testing.T) {
	fs := NewMemFs()
	err := runInit(fs)
	require.NoError(t, err)

	exists, err := fs.Exists("GitVersion.yml")
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := fs.ReadFile("GitVersion.yml")
	require.NoError(t, err)

	expectedContent := `next-version: 0.1.0
major-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(breaking|major))"
minor-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(feature|minor))"
patch-version-bump-message: "^(\\s|\\S)*?(\\+semver:\\s?(fix|patch))"
branches:
  main:
    tag: ""
  develop:
    mode: ContinuousDeployment
    tag: alpha
  release/*:
    mode: semver-from-branch
    tag: beta
    is-release-branch: true
  feature/*:
    tag: use-branch-name
  hotfix/*:
    mode: semver-from-branch
    tag: beta
    is-release-branch: true
`
	assert.Equal(t, expectedContent, string(content))
}

func TestRunInit_FileAlreadyExists(t *testing.T) {
	fs := NewMemFs()
	_, err := fs.Create("GitVersion.yml")
	require.NoError(t, err)

	err = runInit(fs)
	require.NoError(t, err)

	// You might want to assert that the file content remains unchanged
	// or that a specific message is printed to the console.
}
