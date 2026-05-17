# Examples

## Minimal public template

```yaml
template: owner/template-repo
settings:
  has_issues: true
  has_discussions: true
  has_wiki: false
  visibility: public
topics:
  - template
  - starter
```

Create a repository:

```sh
gh template create my-project --manifest ./template-metadata.yml
```

## Locked-down personal project

```yaml
template: owner/go-cli-template
settings:
  has_issues: true
  has_projects: false
  has_discussions: false
  has_pull_requests: true
  pull_request_creation_policy: collaborators_only
  allow_squash_merge: true
  allow_merge_commit: false
  allow_rebase_merge: true
  allow_auto_merge: true
  delete_branch_on_merge: true
  allow_update_branch: true
  visibility: private
actions:
  can_approve_pull_request_reviews: false
  default_workflow_permissions: read
security:
  dependabot_alerts: true
  dependabot_security_updates: true
```

Create a private repository:

```sh
gh template create private-tool --manifest ./template-metadata.yml --private
```

## Production deployment environment

```yaml
template: owner/service-template
settings:
  has_issues: true
  has_pull_requests: true
  delete_branch_on_merge: true
topics:
  - service
  - production
environments:
  - name: production
    wait_timer: 10
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
  - name: SERVICE_TIER
    value: production
secrets:
  - name: DEPLOY_TOKEN
    value: PLACEHOLDER
```

After creating or syncing, replace `PLACEHOLDER` secrets in GitHub.

## Drift check in CI

Use JSON output to make drift checks scriptable:

```sh
gh template audit --repo owner/repo --manifest ./template-metadata.yml --format json
```

The command exits with code 0 when no drift is found and 1 when drift is found.

## Publish a template manifest

Template maintainers can commit `template-metadata.yml` at the root of a template
repository. Users can then create in one step:

```sh
gh template create new-service --manifest owner/service-template
```

Or fetch first for review:

```sh
gh template fetch --repo owner/service-template --output ./template-metadata.yml
gh template create new-service --manifest ./template-metadata.yml
```

## Refresh a manifest from a known-good repository

```sh
gh template snapshot --repo owner/golden-repo --output ./template-metadata.yml
git diff -- template-metadata.yml
```

Review the diff before committing the refreshed manifest. A snapshot records
secret names only; it does not contain real secret values.
