# gh-template
A GitHub CLI extension to snapshot existing repositories' configuration, enhance the experience of provisioning repositories from templates also applying comprehensive configurations, detect configuration drift and apply remediation.

## Commands

| Command | Description |
|---|---|
| `gh template create <name>` | Create a repository from a template and apply a config |
| `gh template snapshot` | Capture a live repository's settings as a YAML manifest |
| `gh template audit` | Detect configuration drift between config and live state |
| `gh template sync` | Reconcile a live repository to match the config |
| `gh template explain [field]` | Show descriptions for all `template-metadata.yml` fields |

## Manifest Reference

The `template-metadata.yml` file drives `create`, `audit`, and `sync`. Run `gh template explain` for terminal-native docs, or consult the table below.

### `settings`

All fields are optional. Omitting a field means it is not applied by `sync` and not checked by `audit`.

| Field | Type | Description |
|---|---|---|
| `has_wiki` | bool | Enable the repository Wiki tab |
| `has_issues` | bool | Enable the Issues tab for bug/feature tracking |
| `has_projects` | bool | Enable the Projects tab (kanban/task boards) |
| `has_discussions` | bool | Enable Discussions for community conversations |
| `has_pull_requests` | bool | Enable the Pull Requests tab |
| `pull_request_creation_policy` | string | Who can open PRs: `"all"` \| `"collaborators_only"` |
| `allow_squash_merge` | bool | Allow squash-merging pull requests |
| `allow_merge_commit` | bool | Allow standard merge commits on pull requests |
| `allow_rebase_merge` | bool | Allow rebase-merging pull requests |
| `allow_auto_merge` | bool | Let PRs auto-merge once required checks pass |
| `delete_branch_on_merge` | bool | Auto-delete source branch after PR merge |
| `allow_update_branch` | bool | Show "Update branch" button on PRs behind base |
| `visibility` | string | Repository visibility: `"public"` or `"private"` |
| `description` | string | Short description shown on the repository page |

### `topics`

A flat list of strings that label the repository on GitHub (e.g. `["go", "cli"]`).

### `environments`

Each entry in the `environments` list configures one GitHub Actions environment.

| Field | Type | Description |
|---|---|---|
| `name` | string | Environment name (e.g. `"production"`, `"staging"`) |
| `wait_timer` | int | Minutes to wait before a deployment can proceed (0–43200) |
| `prevent_self_review` | bool | Prevent the deployer from approving their own deployment |
| `reviewers` | []string | Usernames or `org/team-slug` references that must approve (up to 6) |
| `deployment_branch_policy` | string | Branch/tag restriction mode: `"all"` \| `"protected"` \| `"custom"` |
| `deployment_branch_patterns` | []string | Glob patterns allowed to deploy (only when policy is `"custom"`) |
| `variables` | []object `{name, value}` | Plaintext environment variables injected into workflow jobs |
| `secrets` | []object `{name, value}` | Named secrets to ensure exist; `value` is always `"PLACEHOLDER"` |

### Example

```yaml
settings:
  has_wiki: false
  has_issues: true
  delete_branch_on_merge: true
  allow_update_branch: true
topics:
  - personal-project
  - go-backend
environments:
  - name: production
    wait_timer: 5
    prevent_self_review: true
    reviewers:
      - rpothin
      - my-org/platform-team
    deployment_branch_policy: custom
    deployment_branch_patterns:
      - main
      - "release/*"
    variables:
      - name: DEPLOY_ENV
        value: production
    secrets:
      - name: API_KEY
        value: "PLACEHOLDER"  # replace after creation
```
