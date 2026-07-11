[![Go Reference](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/goloop/cli) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](https://github.com/goloop/cli/blob/master/LICENSE) [![Stay with Ukraine](https://img.shields.io/static/v1?label=Stay%20with&message=Ukraine%20♥&color=ffD700&labelColor=0057B8&style=flat)](https://u24.gov.ua/)

# cli

`goloop` is a small command-line tool to scaffold and inspect goloop-first Go
projects. It is a plain dispatcher over `os.Args` with no third-party
dependencies, and it never commits, pushes or deploys.

## Install

```bash
go install github.com/goloop/cli@latest
```

The binary is named `goloop`.

## Commands

```text
goloop new <name> [--module path] [--dir path] [--force]
goloop check [dir]     scan for AI-assistant traces (exit non-zero on any)
goloop doctor          check that expected tools are installed
goloop status          summarize the module in the current directory
goloop version         print the CLI version
```

### new

Scaffolds a minimal goloop-first service (a `main.go` built on
[`goloop/app`](https://github.com/goloop/app), plus `go.mod`, `.env`,
`.gitignore` and `README.md`). The name must be lowercase letters, digits and
single dashes. It refuses to overwrite an existing directory without `--force`.

### check

Walks a directory and reports any AI-assistant traces (skipping `.git`,
`node_modules`, `vendor`, `dist`). Use it in CI to keep product repositories
clean. Exits non-zero when anything is found.

### doctor

Reports which expected tools (`go`, `git`, `docker`, `node`) are installed.

## Scope

The MVP covers `new`, `check`, `doctor`, `status` and `version`. Richer project
orchestration (multi-repo topologies, deploy scaffolding, compose smoke tests)
is intentionally out of the first version.

## Documentation

- English reference: [DOC.md](DOC.md)
- Ukrainian reference: [DOC.UK.md](DOC.UK.md)

## License

MIT - see [LICENSE](LICENSE).
