# GitVersion Go

[![CI](https://github.com/DBirth/gitversion-go/actions/workflows/ci.yml/badge.svg)](https://github.com/DBirth/gitversion-go/actions/workflows/ci.yml)
[![Security & License Compliance](https://github.com/DBirth/gitversion-go/actions/workflows/security.yml/badge.svg)](https://github.com/DBirth/gitversion-go/actions/workflows/security.yml)
[![Release](https://github.com/DBirth/gitversion-go/actions/workflows/release.yml/badge.svg)](https://github.com/DBirth/gitversion-go/actions/workflows/release.yml)

A command-line tool, written in Go, to help you with semantic versioning based on your Git history and Conventional Commits.

## Features

-   **Configurable Strategies**: Define your versioning workflow with a flexible, ordered list of strategies (e.g., `find-latest-tag`, `increment-from-commits`).
-   **Conventional Commits Support**: Automatically determines version bumps from Conventional Commit messages (`feat:`, `fix:`, `BREAKING CHANGE:`). If the latest commit contains `+semver: none` or `+semver: skip`, no bump is applied (this always takes precedence).
-   **Customizable Commit Bumps**: Use regex to define your own commit message conventions for version bumping (e.g., `+semver: major`). Both conventional and custom bump messages are supported.
-   **Branch-Based Configuration**: Highly configurable versioning rules for different types of branches (e.g., `main`, `develop`, `feature`, `release`).
-   **Prerelease Tags**: Automatically generates prerelease tags (e.g., `-alpha.1`, `-beta.3`, `-feature-new-stuff.5`) based on branch configuration.
-   **Merge Commit Detection**: Supports customizable `merge-message-formats` for advanced merge commit detection.
-   **Commit Date Formatting**: Supports `commit-date-format` for custom commit date output.
-   **Ignore SHAs**: Allows ignoring commits by SHA for version calculation.
-   **Simple Setup**: Get started quickly with a single `GitVersion.yml` configuration file.
-   **Full Test Compliance**: Passes a comprehensive test suite for all supported features and workflows.
-   **Parity with Original GitVersion**: All major features from the original GitVersion are supported and documented.


## Installation

To install the tool, you can use `go install`:

```sh
go install gitversion-go
```

## Test Compliance and Parity

- All features are covered by a comprehensive test suite, including edge cases for tag prefix, no-bump-message, increment settings, ignored SHAs, and conventional commits.
- The tool is fully compliant with GitFlow and GitHubFlow workflows, and matches the behavior of the original GitVersion for all supported features.
- See `parity.md` for a detailed feature parity table.

## Troubleshooting

- **Tag Prefix:** If your tags are not being recognized, ensure your `tag-prefix` matches your tag style. The default (`([vV])?`) matches both `v1.2.3` and `1.2.3`.
- **No-Bump Message:** If you want to suppress a version bump, ensure the latest commit contains `+semver: none` or `+semver: skip`.
- **Custom Bump Patterns:** You can override the default bump regexes for major/minor/patch in your config if you use a different convention.

## Usage

The tool has two main commands: `init` and `calculate`.

### `init`

This command creates a default `GitVersion.yml` configuration file in the current directory. You can specify a workflow template:

```sh
gitversion-go init --workflow GitFlow      # Classic GitFlow branches

gitversion-go init --workflow GitHubFlow   # Simple GitHubFlow (main + feature/*)
```

If no workflow is specified, `GitFlow` is used by default.

### `calculate`

This command analyzes the Git history, reads the `GitVersion.yml` configuration, and calculates the next version. It outputs the version variables in JSON format.

- **If the latest commit contains `+semver: none` or `+semver: skip`, the version is NOT bumped, regardless of other commit messages.**
- **Both `v1.2.3` and `1.2.3` tags are recognized as valid version tags by default.**

```sh
gitversion-go calculate
```

**Example Output:**

```json
{
  "Major": "2",
  "Minor": "0",
  "Patch": "0",
  "PreReleaseTag": "",
  "FullSemVer": "2.0.0"
}
```

## Configuration

Versioning behavior is controlled by a `GitVersion.yml` file. You can generate a ready-made config for your workflow using `gitversion-go init --workflow <GitFlow|GitHubFlow>`.

### Key Options

- `tag-prefix`: Default is `([vV])?` (matches both `v1.2.3` and `1.2.3`).
- `no-bump-message`: If the latest commit contains `+semver: none` or `+semver: skip`, no bump is ever applied (takes precedence over all other rules).
- `commit-date-format`: Go time format string for commit dates (default: ISO8601 `2006-01-02T15:04:05Z07:00`).
- `merge-message-formats`: List of regex patterns to detect merge commits (defaults to common GitHub/GitLab/Bitbucket patterns).
- `ignore`: List of commit SHAs to ignore when calculating bumps.
- `major-version-bump-message`, `minor-version-bump-message`, `patch-version-bump-message`: Regexes for custom bump detection.
- `branches`: Highly configurable branch-based rules.

You can now customize commit date formatting and merge commit detection:

- `commit-date-format`: Go time format string for commit dates (default: ISO8601 `2006-01-02T15:04:05Z07:00`).
- `merge-message-formats`: List of regex patterns to detect merge commits (defaults to common GitHub/GitLab/Bitbucket patterns).
- `tag-prefix`: Regex to match version tags. Default is `([vV])?` so both `v1.0.0` and `1.0.0` tags are recognized.
- `no-bump-message`: Regex for messages that suppress version bumping. If the latest commit contains `+semver: none` or `+semver: skip`, version is never bumped (even if it matches a bump pattern).
- `ignore`: List of commit SHAs to ignore when calculating bumps.

Example:
```yaml
commit-date-format: "2006-01-02 15:04:05"
merge-message-formats:
  - "^Merge pull request #"
  - "^Merge branch '"
  - "^Merged in "
```

The core of the configuration is the `strategies` block, which defines a sequence of versioning strategies to be executed.

GitVersion will try each strategy in order until one successfully determines the version.

### Versioning Strategies

You can define a list of strategies globally or per-branch. The following strategies are available:

-   **`find-latest-tag`**: This strategy finds the latest semantic version tag in the repository's history. It acts as the base version for subsequent strategies.
-   **`increment-from-commits`**: This strategy inspects commit messages since the last tag. It uses **Conventional Commits** (`feat:`, `fix:`, `feat!:`, `BREAKING CHANGE:`) and configurable regex patterns to determine the version bump (`major`, `minor`, or `patch`).
-   **`configured-next-version`**: This strategy acts as a fallback. If no tags are found, it uses the version specified in the `next-version` field of your configuration.

### Example Workflow Templates

You can generate a starter config for either workflow using the CLI:

#### GitFlow
```sh
gitversion-go init --workflow GitFlow
```
Produces:
```yaml
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
```

#### GitHubFlow
```sh
gitversion-go init --workflow GitHubFlow
```
Produces:
```yaml
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
```

## Credits and Disclaimer

This project is a Go implementation of the core concepts from the original [.NET-based GitVersion](https://github.com/GitTools/GitVersion), created by the talented team at GitTools and its contributors. A huge thank you to them for their foundational work in making semantic versioning from Git history so accessible.

This Go version was developed as an experiment in AI-assisted programming to explore the capabilities of agentic AI in refactoring and building software.

For a detailed comparison of features between this project and the original GitVersion, please see the [Feature Parity Document](parity.md).

The original GitVersion project is licensed under the MIT License, and its copyright notice is included below as required:

```
The MIT License (MIT)

Copyright (c) 2021 NServiceBus Ltd, GitTools and contributors.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
