package gitversion

// GetWorkflowTemplate returns a YAML template for the given workflow name.
func GetWorkflowTemplate(name string) string {
	switch name {
	case "GitFlow":
		return `# GitFlow workflow configuration for GitVersion
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
