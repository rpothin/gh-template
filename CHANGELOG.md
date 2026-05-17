# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-17

### Added

- All 9 commands: `create`, `snapshot`, `audit`, `sync`, `list`, `search`,
  `fetch`, `explain`, and `completion`.
- Full manifest schema: `settings`, `topics`, `environments`, `variables`,
  `secrets`, `actions`, and `security`.
- `--format table|json` for `list`, `search`, and `audit`.
- `--include-archived` for `list` and `search`.
- Strict YAML parsing and enum validation.
