package main

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"
	"gitversion-go/internal/fs"
)

func TestInitCommand_GitFlow(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	fsys := fs.NewOsFs()
	err := runInit(fsys, "GitFlow")
	require.NoError(t, err)

	data, err := os.ReadFile("GitVersion.yml")
	require.NoError(t, err)
	content := string(data)
	require.Contains(t, content, "^feature/.*$:")
	require.Contains(t, content, "mode: ContinuousDeployment")
}

func TestInitCommand_GitHubFlow(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	fsys := fs.NewOsFs()
	err := runInit(fsys, "GitHubFlow")
	require.NoError(t, err)

	data, err := os.ReadFile("GitVersion.yml")
	require.NoError(t, err)
	content := string(data)
	require.Contains(t, content, "^main$:")
	require.Contains(t, content, "mode: ContinuousDeployment")
}

func TestInitCommand_UnknownWorkflow(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	fsys := fs.NewOsFs()
	err := runInit(fsys, "UnknownFlow")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown workflow")
}
