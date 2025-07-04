package tests

import (
	"os"
	"testing"

	"gitversion-go/internal/app"
	"gitversion-go/internal/fs"

	"github.com/stretchr/testify/require"
)

func TestInitCommand_GitFlow(t *testing.T) {
	var err error
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		if err = os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to chdir back: %v", err)
		}
	}()
	err = os.Chdir(dir)
	require.NoError(t, err)

	fsys := fs.NewOsFs()
	err = app.RunInit(fsys, "GitFlow")
	require.NoError(t, err)

	data, err := os.ReadFile("GitVersion.yml")
	require.NoError(t, err)
	content := string(data)
	require.Contains(t, content, "^feature/.*$:")
	require.Contains(t, content, "mode: ContinuousDeployment")
}

func TestInitCommand_GitHubFlow(t *testing.T) {
	var err error
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		if err = os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to chdir back: %v", err)
		}
	}()
	err = os.Chdir(dir)
	require.NoError(t, err)

	fsys := fs.NewOsFs()
	err = app.RunInit(fsys, "GitHubFlow")
	require.NoError(t, err)

	data, err := os.ReadFile("GitVersion.yml")
	require.NoError(t, err)
	content := string(data)
	require.Contains(t, content, "^main$:")
	require.Contains(t, content, "mode: ContinuousDeployment")
}

func TestInitCommand_UnknownWorkflow(t *testing.T) {
	var err error
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		if err = os.Chdir(oldWd); err != nil {
			t.Fatalf("failed to chdir back: %v", err)
		}
	}()
	err = os.Chdir(dir)
	require.NoError(t, err)

	fsys := fs.NewOsFs()
	err = app.RunInit(fsys, "UnknownFlow")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown workflow")
}
