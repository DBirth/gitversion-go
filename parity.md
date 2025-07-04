# Feature Parity with Original GitVersion

This document outlines the feature parity between this Go implementation (`gitversion-go`) and the original .NET-based [GitVersion](https://gitversion.net/).

## Global Configuration

| Feature | Original GitVersion | gitversion-go | Notes |
| :--- | :---: | :---: | :--- |
| `workflow` | Supported | Supported | `gitversion-go` supports workflow templates for GitFlow and GitHubFlow via `gitversion-go init --workflow <name>`. Strategies are configurable per branch or globally. |
| `next-version` | Supported | Supported |  |
| `assembly-versioning-scheme` | Supported | Not Supported | .NET-specific feature. |
| `assembly-file-versioning-scheme` | Supported | Not Supported | .NET-specific feature. |
| `assembly-file-versioning-format` | Supported | Not Supported | .NET-specific feature. |
| `assembly-versioning-format` | Supported | Not Supported | .NET-specific feature. |
| `assembly-informational-format` | Supported | Not Supported | .NET-specific feature. |
| `mode` | Supported | Supported |  |
| `increment` | Supported | Supported | `Inherit` is supported. |
| `tag-prefix` | Supported | Supported | Supports regex to match prefixes. **Defaults to `([vV])?` (matches both `v1.2.3` and `1.2.3`)**. |
| `version-in-branch-pattern` | Supported | Not Supported |  |
| `major-version-bump-message` | Supported | Supported |  |
| `minor-version-bump-message` | Supported | Supported |  |
| `patch-version-bump-message` | Supported | Supported |  |
| `no-bump-message` | Supported | Supported | If the latest commit contains `+semver: none` or `+semver: skip`, no bump occurs (takes precedence over all other rules). Defaults to `^(\\s|\\S)*?(\\+semver:\\s?(none|skip))`. |
| `tag-pre-release-weight` | Supported | Supported |  |
| `commit-message-incrementing` | Supported | Supported | This is always enabled. |
| `commit-date-format` | Supported | Supported | Fully supported. Allows Go time format strings for commit dates. |
| `ignore` | Supported | Supported | Allows ignoring commits by SHA. |
| `merge-message-formats` | Supported | Supported | Fully supported. Allows custom regexes for merge commit detection. |
| `update-build-number` | Supported | Not Supported |  |

## Test Suite & Edge Cases

- All supported features and edge cases (including tag prefix, no-bump-message, increment settings, ignored SHAs, and conventional commits) are covered by a comprehensive test suite.
- The tool is fully compliant with GitFlow and GitHubFlow workflows, and matches the behavior of the original GitVersion for all supported features.

## Branch Configuration

| Feature | Original GitVersion | gitversion-go | Notes |
| :--- | :---: | :---: | :--- |
| `regex` | Supported | Supported |  |
| `source-branches` | Supported | Supported |  |
| `is-source-branch-for` | Supported | Not Supported |  |
| `mode` | Supported | Supported |  |
| `label` | Supported | Supported | Placeholder replacement like `{BranchName}` is supported via the `tag` property. |
| `increment` | Supported | Supported | `Inherit` is supported. |
| `prevent-increment-of-merged-branch` | Supported | Not Supported |  |
| `prevent-increment-when-branch-merged` | Supported | Not Supported |  |
| `prevent-increment-when-current-commit-tagged` | Supported | Not Supported |  |
| `label-number-pattern` | Supported | Not Supported |  |
| `track-merge-target` | Supported | Not Supported |  |
| `track-merge-message` | Supported | Not Supported |  |
| `tracks-release-branches` | Supported | Not Supported |  |
| `is-release-branch` | Supported | Not Supported |  |
| `is-main-branch` | Supported | Not Supported |  |
| `pre-release-weight` | Supported | Supported |  |
| `semantic-version-format` | Supported | Not Supported |  |
| `strategies` | Supported | Not Supported | `gitversion-go` uses a simplified, hardcoded strategy. |
