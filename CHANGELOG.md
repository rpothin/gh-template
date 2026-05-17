# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-05-17

### Added

- Two new security manifest fields: `private_vulnerability_reporting` (toggles
  private vulnerability reporting via the GitHub API) and `dependency_graph`
  (toggles the dependency graph; always enabled for public repos).

### Changed

- `sync` and `audit` now treat `description` and `topics` as **seed-only**
  fields: they are applied once at repo creation and left untouched once the
  repo owner has customised them.
- Updated root command help text to better explain the tool's purpose and its
  three core workflows (`create`, `audit`, `sync`).

### Fixed

- Removed the placeholder `value` field from secrets in manifest output
  (`snapshot`). Secret values cannot be read from the GitHub API; the field
  was misleading and is now omitted.

## [0.1.0] - 2026-05-17

### Added

- All 9 commands: `create`, `snapshot`, `audit`, `sync`, `list`, `search`,
  `fetch`, `explain`, and `completion`.
- Full manifest schema: `settings`, `topics`, `environments`, `variables`,
  `secrets`, `actions`, and `security`.
- `--format table|json` for `list`, `search`, and `audit`.
- `--include-archived` for `list` and `search`.
- Strict YAML parsing and enum validation.

[Unreleased]: https://github.com/rpothin/gh-template/compare/v0.1.1...HEAD
[0.1.0]: https://github.com/rpothin/gh-template/releases/tag/v0.1.0
[0.1.1]: https://github.com/rpothin/gh-template/releases/tag/v0.1.1
