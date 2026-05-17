# Getting started

This walkthrough captures a repository's current configuration, reviews the
generated manifest, creates a new repository from that manifest, and checks for
configuration drift.

## 1. Snapshot a source repository

```sh
gh template snapshot --repo owner/source-repo --output ./template-metadata.yml
```

The snapshot includes repository settings, topics, environments, repository
variables, repository secret names, Actions permissions, and supported security
settings. Secret values are never read from GitHub; secret entries use the
literal value `PLACEHOLDER`.

If `--output` is omitted, the YAML is printed to stdout.

## 2. Review the manifest

Open `template-metadata.yml` and remove anything you do not want to enforce on
future repositories. Omitted optional fields are not applied by `sync` and are
not checked by `audit`.

At minimum, `create` needs to know which template repository to use. A snapshot
sets the top-level `template` field automatically:

```yaml
template: owner/source-repo
```

## 3. Create a repository

```sh
gh template create my-new-repo --manifest ./template-metadata.yml
```

`create` creates the repository from the manifest's template repository, then
applies the configured settings, topics, environments, Actions permissions,
variables, secrets, and security settings.

Repository visibility is controlled by the `settings.visibility` field in the
manifest. If the field is omitted, the repository is created as public.

To create the repository under a specific owner (user or organisation), use the
`OWNER/REPO` argument format:

```sh
gh template create owner/my-new-repo --manifest ./template-metadata.yml
```

You can also point `--manifest` at a repository that contains
`template-metadata.yml` at its root:

```sh
gh template create my-new-repo --manifest owner/template-repo
```

## 4. Replace placeholder secrets

After creation, replace any `PLACEHOLDER` secret values in GitHub. The extension
can ensure a secret exists, but GitHub does not allow tools to read back the real
secret value.

## 5. Audit drift

```sh
gh template audit --repo owner/my-new-repo --manifest ./template-metadata.yml
```

`audit` exits with code 0 when no drift is found. It exits with code 1 when
drift is found or when the command cannot complete.

For CI-friendly output:

```sh
gh template audit --repo owner/my-new-repo --manifest ./template-metadata.yml --format json
```

## 6. Sync drift

```sh
gh template sync --repo owner/my-new-repo --manifest ./template-metadata.yml
```

`sync` applies the manifest to an existing repository. Review the manifest before
running it because configured values become the desired state for supported
settings.

## 7. Discover reusable templates

List templates you own:

```sh
gh template list
```

Include organization templates:

```sh
gh template list --include-orgs
```

Search public templates:

```sh
gh template search go cli
```

Fetch a published manifest for local review:

```sh
gh template fetch --repo owner/template-repo --output ./template-metadata.yml
```
