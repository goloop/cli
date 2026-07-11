# goloop CLI - reference

`goloop` scaffolds and inspects goloop-first Go projects. Full English
reference; Ukrainian in [DOC.UK.md](DOC.UK.md).

## Contents

- [Design](#design)
- [new](#new)
- [check](#check)
- [doctor](#doctor)
- [status](#status)
- [version](#version)
- [Scope](#scope)

## Design

The tool is a small dispatcher over `os.Args[1]`: one subcommand per top-level
argument, with per-command flags parsed inline. There is no third-party
dependency and no hidden state. It orchestrates the toolchain by running your
existing tools; it never commits, pushes or deploys.

## new

```text
goloop new <name> [--module path] [--dir path] [--force]
```

Scaffolds a minimal goloop-first service:

| File | Purpose |
|------|---------|
| `main.go` | a `run()` service built on `goloop/app` |
| `go.mod` | module set to `--module` (default `github.com/example/<name>`) |
| `.env` | starter configuration |
| `.gitignore` | build-artifact ignores |
| `README.md` | how to run |

The name must match `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` and must not be a reserved
word. Without `--force`, an existing target directory is left untouched.

## check

```text
goloop check [dir]
```

Walks `dir` (default `.`) and reports every line that contains an AI-assistant
marker, skipping `.git`, `node_modules`, `vendor` and `dist`. It exits `0` when
clean and `1` when any marker is found, so it slots into CI to keep a product
repository free of assistant traces. Unreadable or binary files are skipped, not
treated as errors.

## doctor

```text
goloop doctor
```

Checks for `go` and `git` (required) and `docker` and `node` (optional),
printing each tool's path or a missing note. Exits non-zero when a required tool
is absent.

## status

```text
goloop status
```

Reads `go.mod` in the current directory and prints the module path, Go version
and any required `github.com/goloop/*` dependencies. It understands both the
single-line (`require path v1`) and block (`require ( ... )`) forms, and ignores
`replace` directives (which redirect a dependency rather than declare one).

## version

```text
goloop version
```

Prints the CLI version and platform. Override the version at build time with
`-ldflags "-X main.Version=v1.2.3"`.

## Scope

The MVP is `new`, `check`, `doctor`, `status` and `version`. Multi-repo project
topologies, deploy scaffolding and container smoke tests are intentionally out
of the first version.
