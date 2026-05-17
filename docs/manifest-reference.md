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
```

## Top-level fields

| Field | Type | Required | Description |
|---|---|---:|---|
| `template` | string | Required by local-file `create` | Source template repository in `owner/repo` format. `snapshot` populates it automatically. |
| `settings` | object | No | Repository settings. |
| `topics` | list of strings | No | Repository topics to set. |
| `environments` | list of objects | No | GitHub Actions deployment environments. |
| `variables` | list of objects | No | Repository-level Actions variables. |
| `secrets` | list of objects | No | Repository-level Actions secrets by name. |
| `actions` | object | No | GitHub Actions repository permissions. |
| `security` | object | No | Supported code security and analysis settings. |

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

Secret scanning fields are always available for public repositories. Private
repositories may require GitHub Advanced Security.
