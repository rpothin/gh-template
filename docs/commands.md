# Commands

Run `gh template --help` or `gh template <command> --help` for terminal-native
help. This page lists every command and flag.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Command completed successfully. For `audit`, no drift was detected. |
| 1 | Command failed. For `audit`, this also means drift was detected. |

## `create`

Create a repository from a template manifest.

```sh
gh template create <name-or-owner/repo> [flags]
```

The argument can be a plain repository name (uses the authenticated user as
owner) or an `OWNER/REPO` reference to create the repository under a specific
owner (user or organisation).

Repository visibility is controlled by the `settings.visibility` field in the
manifest. If the field is omitted, the repository is created as public.

Flags:

| Flag | Default | Description |
|---|---|---|
| `-m, --manifest <path-or-owner/repo>` | `./template-metadata.yml` | Local manifest path or repository reference containing `template-metadata.yml`. |

Examples:

```sh
gh template create my-new-repo
gh template create my-new-repo --manifest ./template-metadata.yml
gh template create owner/my-new-repo --manifest ./template-metadata.yml
gh template create my-new-repo --manifest owner/template-repo
```

## `snapshot`

Snapshot a repository's settings into YAML.

```sh
gh template snapshot --repo owner/repo [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-r, --repo <owner/repo>` | Required | Repository to snapshot. |
| `-o, --output <path>` | stdout | File path to write YAML to. If the path is an existing directory, writes `template-metadata.yml` inside it. |

Examples:

```sh
gh template snapshot --repo owner/repo
gh template snapshot --repo owner/repo --output ./template-metadata.yml
gh template snapshot --repo owner/repo --output ./manifests/
```

## `audit`

Audit a repository against a template manifest.

```sh
gh template audit --repo owner/repo [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-r, --repo <owner/repo>` | Required | Repository to audit. |
| `-m, --manifest <path>` | `./template-metadata.yml` | Manifest to compare against live state. |
| `--format <table|json>` | `table` | Output format. |

Examples:

```sh
gh template audit --repo owner/repo
gh template audit --repo owner/repo --manifest ./custom-template.yml
gh template audit --repo owner/repo --format json
```

In JSON mode, the report includes `repo`, `manifest`, `drift_count`, `drifts`,
and `warnings`.

## `sync`

Sync repository settings from a template manifest.

```sh
gh template sync --repo owner/repo [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-r, --repo <owner/repo>` | Required | Repository to update. |
| `-m, --manifest <path>` | `./template-metadata.yml` | Manifest to apply. |
| `-b, --branch <branch>` | `chore/sync-common-files` | Branch for common-file commits. |

Examples:

```sh
gh template sync --repo owner/repo
gh template sync --repo owner/repo --manifest ./custom-template.yml
gh template sync --repo owner/repo --branch feat/update-workflows
gh template sync --repo owner/repo --branch main
```

`sync` applies supported settings from the manifest. Existing repository
variables not present in the manifest are left untouched. Existing secret values
are not read back; missing secrets are initialized with the placeholder value.

If the manifest includes a `common_files` section, the listed files and
directories are copied from the template repository (`manifest.template`) to the
target repository.  By default this is committed to the branch
`chore/sync-common-files` (created automatically from the repository's default
branch if it does not exist), so the changes can be reviewed via a pull request.
Use `--branch` to target a specific branch; use `--branch main` (or the repo's
actual default branch name) to commit directly.

## `list`

List template repositories owned by the authenticated user.

```sh
gh template list [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `--include-orgs` | `false` | Also list template repositories from organizations you belong to. |
| `--include-archived` | `false` | Include archived template repositories. |
| `--format <table|json>` | `table` | Output format. |

Examples:

```sh
gh template list
gh template list --include-orgs
gh template list --include-archived
gh template list --format json
```

## `search`

Search for public template repositories.

```sh
gh template search [query] [flags]
```

The query is automatically prefixed with `template:true`. Results are sorted by
star count descending.

Flags:

| Flag | Default | Description |
|---|---|---|
| `--limit <n>` | `30` | Maximum number of results to return. The maximum is 100. |
| `--include-archived` | `false` | Include archived template repositories. |
| `--format <table|json>` | `table` | Output format. |

Examples:

```sh
gh template search
gh template search go cli
gh template search starter language:go
gh template search org:github --limit 50
gh template search go --format json
```

## `fetch`

Fetch `template-metadata.yml` from a repository.

```sh
gh template fetch --repo owner/repo [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `-r, --repo <owner/repo>` | Required | Repository containing `template-metadata.yml` at its root. |
| `-o, --output <path>` | `template-metadata.yml` | File path to write the manifest. |

Examples:

```sh
gh template fetch --repo owner/template-repo
gh template fetch --repo owner/template-repo --output ./manifests/template.yml
```

## `explain`

Show descriptions for `template-metadata.yml` fields.

```sh
gh template explain [field] [flags]
```

Flags:

| Flag | Default | Description |
|---|---|---|
| `--all` | `false` | Show detailed descriptions for every field. |

Examples:

```sh
gh template explain
gh template explain visibility
gh template explain --all
```

## `completion`

Generate shell autocompletion scripts.

```sh
gh template completion [bash|zsh|fish|powershell]
```

Completions register for the binary name `template`, not `gh template`, because
of GitHub CLI extension dispatch. See [Installation](installation.md#shell-completion)
for setup examples.
