# Manifest reference

`template-metadata.yml` drives `create`, `audit`, and `sync`. YAML is parsed
strictly: unknown fields are rejected, and enum values must match the accepted
strings exactly.

All sections are optional unless noted. Omitting an optional field means it is
not applied by `sync` and not checked by `audit`.

## Complete example

```yaml
template: owner/template-repo
settings:
  has_wiki: false
  has_issues: true
  has_projects: true
  has_discussions: true
  has_pull_requests: true
  pull_request_creation_policy: collaborators_only
  allow_squash_merge: true
  allow_merge_commit: false
  allow_rebase_merge: true
  allow_auto_merge: true
  delete_branch_on_merge: true
  allow_update_branch: true
  visibility: public
  description: Example repository created from a manifest
topics:
  - go
  - cli
environments:
  - name: production
    wait_timer: 5
    prevent_self_review: true
    reviewers:
      - owner
      - org/platform-team
    deployment_branch_policy: selected
    deployment_branch_patterns:
      - main
      - "release/*"
    variables:
      - name: DEPLOY_ENV
        value: production
    secrets:
      - name: API_KEY
        value: PLACEHOLDER
variables:
  - name: DEPLOY_URL
    value: https://example.com
secrets:
  - name: API_TOKEN
    value: PLACEHOLDER
actions:
  can_approve_pull_request_reviews: false
  sha_pinning_required: false
  default_workflow_permissions: read
security:
  dependabot_alerts: true
  dependabot_security_updates: true
  secret_scanning: true
  secret_scanning_push_protection: true
common_files:
  - .github/workflows/
  - AGENTS.md
  - docs/skills/
```

## Top-level fields

| Field | Type | Required | Description |
|---|---|---:|---|
| `template` | string | Required by local-file `create` | Source template repository in `owner/repo` format. `snapshot` populates it automatically. Required by `sync` when `common_files` is set. |
| `settings` | object | No | Repository settings. |
| `topics` | list of strings | No | Repository topics to set. |
| `environments` | list of objects | No | GitHub Actions deployment environments. |
| `variables` | list of objects | No | Repository-level Actions variables. |
| `secrets` | list of objects | No | Repository-level Actions secrets by name. |
| `actions` | object | No | GitHub Actions repository permissions. |
| `security` | object | No | Supported code security and analysis settings. |
| `common_files` | list of strings | No | Files or directories to copy from the template repository during `sync` and to check for drift during `audit`. |

When `create --manifest owner/repo` is used, the repository reference is used as
the template fallback if the remote manifest omits `template`.

## `settings`

| Field | Type | Accepted values | Description |
|---|---|---|---|
| `has_wiki` | bool | `true`, `false` | Enable the Wiki tab. |
| `has_issues` | bool | `true`, `false` | Enable the Issues tab. |
| `has_projects` | bool | `true`, `false` | Enable repository Projects. |
| `has_discussions` | bool | `true`, `false` | Enable GitHub Discussions. |
| `has_pull_requests` | bool | `true`, `false` | Enable pull requests. |
| `pull_request_creation_policy` | string | `collaborators_only`, `contributors_only`, `all_users`, `any_user` | Controls who can open pull requests. |
| `allow_squash_merge` | bool | `true`, `false` | Allow squash merges. |
| `allow_merge_commit` | bool | `true`, `false` | Allow merge commits. |
| `allow_rebase_merge` | bool | `true`, `false` | Allow rebase merges. |
| `allow_auto_merge` | bool | `true`, `false` | Allow auto-merge. |
| `delete_branch_on_merge` | bool | `true`, `false` | Delete PR branches after merge. |
| `allow_update_branch` | bool | `true`, `false` | Show the update branch button on PRs. |
| `visibility` | string | `public`, `private`, `internal` | Repository visibility. |
| `description` | string | Any string | Repository description. |

## `topics`

`topics` is a flat list of GitHub topic names:

```yaml
topics:
  - go
  - github-cli
  - template
```

## `environments`

Each item configures one GitHub Actions environment.

| Field | Type | Accepted values | Description |
|---|---|---|---|
| `name` | string | Any environment name | Environment name. |
| `wait_timer` | int | `0` through `43200` on GitHub | Minutes to wait before deployments can proceed. |
| `prevent_self_review` | bool | `true`, `false` | Prevent the deployer from approving their own deployment. |
| `reviewers` | list of strings | GitHub usernames or `org/team-slug` | Required deployment reviewers. GitHub allows up to 6. |
| `deployment_branch_policy` | string | `all`, `protected`, `selected` | Branch or tag restriction mode. |
| `deployment_branch_patterns` | list of strings | Branch or tag patterns | Allowed branch/tag patterns when policy is `selected`. |
| `variables` | list of `{name, value}` | Plaintext values | Environment-level Actions variables. |
| `secrets` | list of `{name, value}` | Secret names with `PLACEHOLDER` values | Environment-level Actions secrets. |

## Repository `variables`

Repository variables are plaintext name/value pairs available to Actions
workflows:

```yaml
variables:
  - name: DEPLOY_URL
    value: https://example.com
```

| Field | Type | Description |
|---|---|---|
| `name` | string | Variable name. |
| `value` | string | Plaintext variable value. |

## Repository `secrets`

Repository secrets are represented by name. Values should be
`PLACEHOLDER` because GitHub secret values are write-only:

```yaml
secrets:
  - name: API_TOKEN
    value: PLACEHOLDER
```

| Field | Type | Description |
|---|---|---|
| `name` | string | Secret name. |
| `value` | string | Placeholder value. Replace the real secret in GitHub after creation. |

## `actions`

| Field | Type | Accepted values | Description |
|---|---|---|---|
| `can_approve_pull_request_reviews` | bool | `true`, `false` | Allow `GITHUB_TOKEN` to create and approve pull requests. |
| `sha_pinning_required` | bool | `true`, `false` | Require Actions references to use full-length commit SHAs. |
| `default_workflow_permissions` | string | `read`, `write` | Default `GITHUB_TOKEN` workflow permission level. |

## `security`

| Field | Type | Accepted values | Description |
|---|---|---|---|
| `dependabot_alerts` | bool | `true`, `false` | Enable Dependabot vulnerability alerts. |
| `dependabot_security_updates` | bool | `true`, `false` | Enable Dependabot security update pull requests. |
| `secret_scanning` | bool | `true`, `false` | Enable secret scanning where available. |
| `secret_scanning_push_protection` | bool | `true`, `false` | Enable push protection where available. |
| `private_vulnerability_reporting` | bool | `true`, `false` | Enable private vulnerability reporting. Primarily available for public repositories; silently treated as unsupported (not an error) for private repositories or organisations where the feature is unavailable. |
| `dependency_graph` | bool | `true`, `false` | Enable the dependency graph. Always enabled for public repositories; setting this to `false` on a public repository may have no effect. |

Secret scanning fields are always available for public repositories. Private
repositories may require GitHub Advanced Security.

### Settings visible in the GitHub UI that cannot be managed via the manifest

The following Advanced Security settings are **not** configurable through the manifest because they require file-based configuration or are controlled by GitHub Actions workflows:

| UI Setting | Reason |
|---|---|
| Automatic dependency submission | Requires the `actions/dependency-review-action` or `advanced-security/dependency-review-toolkit` GitHub Actions workflow. |
| Dependabot rules | Complex rule objects; no simple boolean REST API field. Configure via the Dependabot UI or a `dependabot.yml` file. |
| Dependabot malware alerts | Shares the same API endpoint as standard Dependabot alerts; there is no separate toggle. |
| Grouped security updates | Controlled via the `groups:` key in a `dependabot.yml` file. |
| Dependabot version updates | Requires a `dependabot.yml` configuration file. |
| Code scanning (CodeQL & other tools) | Requires a GitHub Actions workflow or GitHub's default setup. |
| Copilot Autofix suggestions | Tied to code scanning; not a standalone repository setting. |
| Code scanning protection rules | Complex threshold-based settings with no corresponding simple manifest field. |

## `common_files`

`common_files` is a flat list of file paths or directory paths (relative to the
repository root) that `sync` copies from the template repository
(`manifest.template`) to the target repository, and that `audit` checks for
drift.

- **File paths** (e.g. `AGENTS.md`, `.github/CODEOWNERS`) are copied as-is.
- **Directory paths** (e.g. `.github/workflows/`, `docs/skills/`) are expanded
  recursively — every file in the directory tree is included.
- Each file is written to the **same relative path** in the target repository.
- Files whose content has not changed (identical git-blob SHA) are **skipped**
  to avoid empty commits.
- The `template` top-level field is required when `common_files` is set.

`audit` compares each file's git-blob SHA in the template repository against the
same path in the target repository. A missing file or a SHA mismatch is reported
as drift in the **Common Files** section.

```yaml
common_files:
  - .github/workflows/
  - AGENTS.md
  - docs/skills/
  - .github/copilot-instructions.md
```

By default, `sync` commits common files to a dedicated branch named
`chore/sync-common-files` (created from the target repository's default branch
if it does not already exist), so that the changes can be reviewed via a pull
request before being merged.  Use the `--branch` flag to target a different
branch.  Pass the name of the default branch (e.g. `--branch main`) to commit
directly without creating a review branch.
