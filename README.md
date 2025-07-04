# GitVersion Go

A command-line tool, written in Go, to help you with semantic versioning based on your Git history and Conventional Commits.

## Features

-   **Configurable Strategies**: Define your versioning workflow with a flexible, ordered list of strategies (e.g., `find-latest-tag`, `increment-from-commits`).
-   **Conventional Commits Support**: Automatically determines version bumps from Conventional Commit messages (`feat:`, `fix:`, `BREAKING CHANGE:`).
-   **Customizable Commit Bumps**: Use regex to define your own commit message conventions for version bumping (e.g., `+semver: major`).
-   **Branch-Based Configuration**: Highly configurable versioning rules for different types of branches (e.g., `main`, `develop`, `feature`, `release`).
-   **Prerelease Tags**: Automatically generates prerelease tags (e.g., `-alpha.1`, `-beta.3`, `-feature-new-stuff.5`) based on branch configuration.
-   **Simple Setup**: Get started quickly with a single `GitVersion.yml` configuration file.

## Installation

To install the tool, you can use `go install`:

```sh
go install gitversion-go
```

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

### Configuration Options

-   `next-version`: The version to start with if no tags are found in the repository. Used by the `configured-next-version` strategy.
-   `strategies`: A list of versioning strategies to execute in order. This can be defined globally and overridden per-branch.
-   `major-version-bump-message`: A regex pattern to identify commits that should trigger a major version bump. Used by the `increment-from-commits` strategy.
-   `minor-version-bump-message`: A regex pattern to identify commits that should trigger a minor version bump. Used by the `increment-from-commits` strategy.
-   `patch-version-bump-message`: A regex pattern to identify commits that should trigger a patch version bump. Used by the `increment-from-commits` strategy.
-   `branches`: A map of branch configurations. The keys are regex patterns to match against branch names.
    -   `strategies`: (Optional) A list of strategies to use for this branch, overriding the global list.
    -   `tag`: The prerelease tag to use for the branch (e.g., `alpha`, `beta`). If set to `use-branch-name`, the branch name will be sanitized and used.
    -   `prevent-increment`: (Optional) A boolean that, if `true`, prevents the patch version from being incremented for each commit on top of the calculated version. Defaults to `false`.
    -   `source-branches`: (Optional) A list of branches from which the current branch can inherit its version. This is useful for feature branches that are based on a development branch.
    -   `pre-release-weight`: (Optional) An integer value that can be used to order prerelease versions.

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
