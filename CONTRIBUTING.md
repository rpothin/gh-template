# Contributing

Thanks for your interest in improving `gh-template`.

This project currently uses a lightweight contribution process:

- Use [Issues](https://github.com/rpothin/gh-template/issues) for reproducible
  bugs.
- Use [Discussions](https://github.com/rpothin/gh-template/discussions) for
  questions, usage help, and feature ideas.
- Please do not open unsolicited code pull requests. Start with an issue or
  discussion first so the scope and approach can be agreed before implementation.

## Bug reports

When reporting a bug, include:

- The command you ran
- The version of `gh`, your operating system, and how you installed the extension
- The expected behavior
- The actual behavior, including sanitized error output
- A minimal `template-metadata.yml` snippet if the bug is manifest-related

Do not include real tokens, secrets, private repository names, or other sensitive
data in public issues.

## Feature ideas

Open a GitHub Discussion for feature ideas. Include the workflow you want to
support, why the current commands are not enough, and any compatibility concerns
with existing manifests.

## Pull requests

Code pull requests are not accepted without prior discussion. If a maintainer
invites a pull request, keep it focused, update relevant docs, and run:

```sh
go test ./...
```

## Code of conduct

Participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md).
