# Contributing to PortMaster

Thank you for considering a contribution. This document describes how to set
up the project, build and test it, and what standards the code follows.

## Development setup

Requirements:

- Go 1.25 or newer
- Git

Clone the repository and verify the build:

```sh
git clone https://github.com/RichardFlp/portmaster.git
cd portmaster
make build
```

## Building

```sh
make build
```

The binary is written to `bin/portmaster`. To build without make:

```sh
go build -o bin/portmaster ./cmd/portmaster
```

## Testing

```sh
make test
```

or:

```sh
go test ./...
```

Run the race detector on supported platforms:

```sh
go test -race ./...
```

Tests use dynamically allocated ports, so they never depend on a specific
port being free. Platform-specific tests run only on their operating system.

## Formatting and static analysis

The project must pass all of the following before a pull request is merged:

```sh
gofmt -l .   # must print nothing
go vet ./...
go test ./...
```

`make lint` runs `gofmt -l .` and `go vet ./...` together.

## Coding standards

- Write idiomatic Go: clear names, small functions, sensible package layout,
  proper error handling.
- Prefer the Go standard library. Dependencies are kept minimal and must be
  well-maintained and appropriately licensed.
- Keep platform-specific code in files tagged with the operating system
  (for example `ports_windows.go`), behind platform-neutral functions. Do not
  spread operating-system logic through the codebase.
- Do not add comments to source code. Code must be understandable through
  naming, structure, and types. If a piece of code needs explanation, improve
  the code instead. Documentation belongs in `README.md` and the `docs`
  directory.
- Do not add AI functionality, telemetry, analytics, accounts, or any
  cloud-dependent behavior. PortMaster runs locally and offline.
- Do not add unused dependencies, dead code, or speculative abstractions.

## Pull requests

1. Create a branch with a descriptive name.
2. Make focused changes and commit them with clear messages.
3. Run `gofmt -l .`, `go vet ./...`, and `go test ./...` and fix any failures.
4. Open a pull request against `main` and describe the change and how it was
   tested.

The CI workflow runs formatting checks, `go vet`, tests, and a build on
Windows, Linux, and macOS. Pull requests must pass CI.

## Issue reports

When opening an issue, include:

- The operating system and architecture.
- The PortMaster version (`portmaster version`).
- The exact command that was run.
- The expected output and the actual output.
- Any relevant process or port context.

Bug reports and feature requests are welcome. Feature requests should explain
the use case so the maintainers can judge scope.

## Security issues

Do not report security issues in the public issue tracker. See
[SECURITY.md](SECURITY.md) for the reporting process.
