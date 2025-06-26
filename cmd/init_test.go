package cmd

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
  release:
    mode: semver-from-branch
    tag: beta
  feature:
    tag: use-branch-name
  hotfix:
    mode: semver-from-branch
    tag: beta
`
	assert.Equal(t, expectedContent, string(content))
}

func TestRunInit_FileAlreadyExists(t *testing.T) {
	fs := NewMemFs()

	// Create an existing file
	originalContent := "some: existing content"
	_, err := fs.Create("GitVersion.yml")
	require.NoError(t, err)
	err = fs.WriteFile("GitVersion.yml", []byte(originalContent), 0644)
	require.NoError(t, err)

	err = runInit(fs)
	require.NoError(t, err)

	// Check that the file was not overwritten
	currentContent, err := fs.ReadFile("GitVersion.yml")
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(currentContent))
}
