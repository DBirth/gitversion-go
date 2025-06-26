package tests

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConventionalCommits(t *testing.T) {
	testCases := []struct {
		name            string
		commitMessage   string
		expectedVersion string
	}{
		{"Feat", "feat: a new feature", "1.1.0"},
		{"Fix", "fix: a bug fix", "1.0.1"},
		{"BreakingChange", "feat!: a breaking change", "2.0.0"},
		{"BreakingChangeBody", "feat: a feature with a breaking change\n\nBREAKING CHANGE: this is a breaking change", "2.0.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTestRepo(t)
			repo.writeFile("initial.txt", "initial commit")
			initialCommit := repo.commit("initial commit")
			repo.tag("1.0.0", initialCommit)

			repo.writeFile(tc.name+".txt", tc.commitMessage)
			repo.commit(tc.commitMessage)

			cmd := exec.Command(binaryPath, "calculate", "--path", repo.path)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, string(output))

			assert.True(t, strings.Contains(string(output), tc.expectedVersion), "output was: %s", string(output))
		})
	}
}
