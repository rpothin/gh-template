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

| Field | Type | Description |
|---|---|---|
| `name` | string | Environment name (e.g. `"production"`, `"staging"`) |
| `wait_timer` | int | Minutes to wait before a deployment can proceed (0–43200) |

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
    wait_timer: 0
```
