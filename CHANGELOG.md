# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-10-18

Report directory traversal errors instead of marking an unreadable path as clean. The v1 release keeps the existing import path and public API. Future incompatible changes require a new major version.

## [0.2.0] - 2026-07-12

### Added
- `check --all` scans every file fail-closed. Without it, a file that looks
  binary (a NUL byte near its start) is skipped, since a compiled artifact or
  image is not source and only produces scan noise.

### Fixed
- `new --dir` rejects a value that traverses above the working directory (a
  `..` segment), so a caller-supplied directory cannot place the scaffold
  outside the intended location.

### Documentation
- Noted that the compiled `cli` binary contains the joined trace markers, so it
  must never be committed to a product repository (keep it in `.gitignore`).

## [0.1.0]

First release: `new`, `check`, `doctor`, `status` and `version` for scaffolding
and inspecting goloop-first projects, with a fail-closed trace scanner and a
go.mod parser, zero third-party dependencies.
