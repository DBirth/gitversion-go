package gitversion

// DefaultConfig is the default configuration for GitVersion.
const DefaultConfig = `next-version: 0.1.0
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
