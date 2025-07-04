package gitversion

// GetWorkflowTemplate returns a YAML template for the given workflow name.
func GetWorkflowTemplate(name string) string {
	switch name {
	case "GitFlow":
		return `# GitFlow workflow configuration for GitVersion
# commit-date-format: "2006-01-02T15:04:05Z07:00" # optional, Go time format string
# merge-message-formats:
#   - "^Merge pull request #"
#   - "^Merge branch '"
#   - "^Merged in "
branches:
  ^master$:
    mode: ContinuousDeployment
    tag: ''
    increment: Patch
    is-release-branch: true
  ^develop$:
    mode: ContinuousDeployment
    tag: beta
    increment: Minor
  ^release/.*$:
    mode: ContinuousDeployment
    tag: rc
    increment: Patch
    is-release-branch: true
  ^hotfix/.*$:
    mode: ContinuousDeployment
    tag: hotfix
    increment: Patch
    is-release-branch: true
  ^feature/.*$:
    mode: ContinuousDeployment
    tag: use-branch-name
    increment: Minor
    source-branches: [develop]
`
	case "GitHubFlow":
		return `# GitHubFlow workflow configuration for GitVersion
# commit-date-format: "2006-01-02T15:04:05Z07:00" # optional, Go time format string
# merge-message-formats:
#   - "^Merge pull request #"
#   - "^Merge branch '"
#   - "^Merged in "
branches:
  ^main$:
    mode: ContinuousDeployment
    tag: ''
    increment: Patch
    is-release-branch: true
  ^feature/.*$:
    mode: ContinuousDeployment
    tag: use-branch-name
    increment: Minor
    source-branches: [main]
`
	default:
		return ""
	}
}
