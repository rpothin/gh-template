# gh-template

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/rpothin/gh-template)](go.mod)

A GitHub CLI extension to snapshot repository configuration, create repositories
from template manifests, audit drift, and sync live repositories back to a known
configuration.

`gh-template` helps individual developers and maintainers avoid ClickOps by
capturing repository settings, topics, GitHub Actions environments, variables,
secrets, Actions permissions, and security settings in `template-metadata.yml`.

## Installation

Prerequisites:

- [GitHub CLI](https://cli.github.com/) authenticated with `gh auth login`
- Git, for installing GitHub CLI extensions

Install the extension:

```sh
gh extension install rpothin/gh-template
gh template --help
```

For source builds and troubleshooting, see [Installation](docs/installation.md).

## Quick start

Capture a repository as a reusable manifest:

```sh
gh template snapshot --repo owner/source-repo --output ./template-metadata.yml
```

Review and customize the manifest, then create a new repository from it:

```sh
gh template create my-new-repo --manifest ./template-metadata.yml
```

Audit or reconcile an existing repository:

```sh
gh template audit --repo owner/target-repo --manifest ./template-metadata.yml
gh template sync --repo owner/target-repo --manifest ./template-metadata.yml
```

See the full walkthrough in [Getting started](docs/getting-started.md).

## Commands

| Command | Description |
|---|---|
| `gh template create <name>` | Create a repository from a template manifest |
| `gh template snapshot` | Capture a live repository's settings as YAML |
| `gh template audit` | Detect configuration drift between a manifest and live state |
| `gh template sync` | Reconcile a live repository to match a manifest |
| `gh template list` | List template repositories you own, with optional org templates |
| `gh template search [query]` | Search public template repositories on GitHub |
| `gh template fetch` | Fetch a template's recommended manifest locally |
| `gh template explain [field]` | Show descriptions for manifest fields |
| `gh template completion [shell]` | Generate shell completion scripts |

Every command and flag is documented in [Commands](docs/commands.md).

## Manifest

The `template-metadata.yml` manifest is the source of truth for `create`,
`audit`, and `sync`. It supports:

- Repository settings and visibility
- Topics
- GitHub Actions environments, reviewers, branch policies, variables, and secrets
- Repository-level Actions variables and secrets
- Actions workflow permissions
- Dependabot and secret-scanning settings

Read [Manifest reference](docs/manifest-reference.md) for the complete schema,
or run:

```sh
gh template explain --all
```

## Examples

Common workflows and manifest examples are available in [Examples](docs/examples.md).

## Contributing

Issues are used for reproducible bugs. Questions and feature ideas belong in
GitHub Discussions. Please read [Contributing](CONTRIBUTING.md) before opening
an issue or proposing changes.

## License

MIT. See [LICENSE](LICENSE).
