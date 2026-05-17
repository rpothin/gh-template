# Installation

## Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated:

  ```sh
  gh auth login
  ```

- Git, because GitHub CLI extensions are installed from Git repositories.

For source builds, install Go 1.26.2 or newer.

## Install as a GitHub CLI extension

```sh
gh extension install rpothin/gh-template
gh template --help
```

Upgrade later with:

```sh
gh extension upgrade gh-template
```

Uninstall with:

```sh
gh extension remove gh-template
```

## Build from source

```sh
git clone https://github.com/rpothin/gh-template.git
cd gh-template
go test ./...
go build -o gh-template .
```

You can run the binary directly:

```sh
./gh-template --help
```

Or install it as a local GitHub CLI extension while developing:

```sh
gh extension install .
gh template --help
```

## Authentication and permissions

Most commands call the GitHub API through the authenticated `gh` session. If a
command fails with a permission error, refresh authentication with scopes that
match the resource you are managing:

```sh
gh auth refresh -h github.com -s repo -s workflow
```

Repository secrets, variables, environments, Actions settings, and security
settings may require owner, admin, or organization-level permissions on the
target repository.

## Shell completion

Generate completion scripts with:

```sh
gh template completion bash
gh template completion zsh
gh template completion fish
gh template completion powershell
```

Due to GitHub CLI extension dispatch, completions register for the executable
name `template`, not for `gh template`. For the best completion experience, add
an alias that points to the extension binary and complete that alias directly.

PowerShell example:

```powershell
Set-Alias -Name template -Value gh-template
gh template completion powershell | Out-String | Invoke-Expression
```

Bash example:

```sh
source <(gh template completion bash)
```

Then use:

```sh
template create --<TAB>
```

## Troubleshooting

### `gh: command not found`

Install GitHub CLI and ensure it is on your `PATH`.

### `error: HTTP 403` or `Resource not accessible by integration`

The authenticated user likely lacks permissions or scopes for the requested
repository setting. Run `gh auth status`, refresh scopes, and confirm your role
on the repository.

### Completion scripts do not work for `gh template`

Use the direct `template` alias described above. This is a limitation of how
GitHub CLI dispatches extensions.
