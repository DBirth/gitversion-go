package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	binaryPath string
)

func TestMain(m *testing.M) {
	var err error
	// Build the gitversion-go binary
	tempDir, err := os.MkdirTemp("", "gitversion-go-tests")
	if err != nil {
		fmt.Printf("Failed to create temp dir for binary: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("Failed to remove temp dir for binary: %v\n", err)
		}
	}()

	binaryPath = filepath.Join(tempDir, "gitversion-go")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/gitversion-go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build gitversion-go binary: %v\nOutput:\n%s\n", err, string(output))
		os.Exit(1)
	}

	// Run the tests
	exitCode := m.Run()

	os.Exit(exitCode)
}
