# gh-template
A GitHub CLI extension to snapshot existing repositories' configuration, enhance the experience of provisioning repositories from templates also applying comprehensive configurations, detect configuration drift and apply remediation.

## Commands

| Command | Description |
|---|---|
| `gh template create <name>` | Create a repository from a template manifest (template repo read from manifest) |
| `gh template snapshot` | Capture a live repository's settings as a YAML manifest |
| `gh template audit` | Detect configuration drift between config and live state |
| `gh template sync` | Reconcile a live repository to match the config |
| `gh template list` | List your own (and org) template repositories |
| `gh template search [query]` | Search public template repositories on GitHub |
| `gh template fetch` | Fetch a template's recommended manifest locally for review |
| `gh template explain [field]` | Show descriptions for all `template-metadata.yml` fields |

## Usage

### Listing and discovering templates

**Your templates** (repos you own marked as templates):
```sh
gh template list
gh template list --include-orgs   # also includes organisation templates
```

**Discover community templates** (public repos on GitHub):
```sh
gh template search go cli
gh template search starter language:go
gh template search                # most starred public templates
gh template search --limit 50     # up to 100 results
```

Once you find a template you want to use, fetch its manifest and create:
```sh
gh template fetch --repo owner/my-template
gh template create my-new-repo --manifest ./template-metadata.yml
# or in one step:
gh template create my-new-repo --manifest owner/my-template
```

### Creating a repository from a template

`gh template create` accepts two forms for `--manifest`:

**Local file** (default — `./template-metadata.yml`):
```sh
gh template create my-new-repo --manifest ./template-metadata.yml
```

**Repository reference** (`owner/repo`): fetches `template-metadata.yml` from the root of the given repository via the GitHub API and applies it in memory:
```sh
gh template create my-new-repo --manifest owner/my-template
```

When a template maintainer ships a `template-metadata.yml` alongside their template repository, this lets you create a fully-configured repository in one command without any local setup.

### Fetching a manifest for local review

To inspect or customise a template's manifest before creating a repository, fetch it locally first:
```sh
gh template fetch --repo owner/my-template
# review / edit ./template-metadata.yml
gh template create my-new-repo --manifest ./template-metadata.yml
```

By default the file is written to `./template-metadata.yml`. Use `--output <path>` to change the destination.

## Manifest Reference

The `template-metadata.yml` file drives `create`, `audit`, and `sync`. Run `gh template explain` for terminal-native docs, or consult the table below.

### Top-level fields

| Field | Type | Description |
|---|---|---|
| `template` | string | Source template repository in `owner/repo` format. Auto-populated by `snapshot`. Required by `create`. |

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
template: rpothin/my-template-repo
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
variables:
  - name: DEPLOY_URL
    value: https://example.com
secrets:
  - name: API_TOKEN
    value: "PLACEHOLDER"  # replace after creation
actions:
  can_approve_pull_request_reviews: false
  sha_pinning_required: false
  default_workflow_permissions: read
security:
  dependabot_alerts: true
  dependabot_security_updates: true
  secret_scanning: true               # public repos or GitHub Advanced Security only
  secret_scanning_push_protection: true
```

### `variables`

Repository-level Actions variables injected into all workflow jobs as environment variables. Values are plaintext and visible in logs — use `secrets` for credentials.

| Field | Type | Description |
|---|---|---|
| `name` | string | Variable name |
| `value` | string | Variable value (plaintext) |

### `secrets`

Repository-level Actions secrets. Because GitHub never returns secret values via the API, `value` is always `"PLACEHOLDER"` in snapshots and manifests.

| Field | Type | Description |
|---|---|---|
| `name` | string | Secret name |
| `value` | string | Always `"PLACEHOLDER"` — update manually after creation |

### `actions`

GitHub Actions workflow permissions for this repository.

| Field | Type | Description |
|---|---|---|
| `can_approve_pull_request_reviews` | bool | Allow the `GITHUB_TOKEN` to create and approve pull requests |
| `sha_pinning_required` | bool | Require Actions to reference a full-length commit SHA (not a mutable tag) |
| `default_workflow_permissions` | string | Default `GITHUB_TOKEN` permission scope: `"read"` \| `"write"` |

### `security`

Code security and analysis settings. Fields unavailable for the repository type are silently ignored by the API.

| Field | Type | Description |
|---|---|---|
| `dependabot_alerts` | bool | Enable Dependabot vulnerability alerts (all repo types) |
| `dependabot_security_updates` | bool | Enable Dependabot automated security-update PRs (requires alerts) |
| `secret_scanning` | bool | Scan for leaked credentials (public repos; or private repos with GitHub Advanced Security) |
| `secret_scanning_push_protection` | bool | Block pushes containing secrets (requires `secret_scanning` enabled) |
