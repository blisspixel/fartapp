# Contributing

F.A.R.T. Lab is a silly idea taken seriously. Small, verified steps are better
than large speculative implementations.

## Before proposing a change

Read:

- [README.md](README.md) for product identity and current status.
- [ROADMAP.md](ROADMAP.md) for sequencing and completion gates.
- [docs/SIMULATION.md](docs/SIMULATION.md) for scientific claims.
- [docs/GAMEPLAY.md](docs/GAMEPLAY.md) for procedural story rules.
- [docs/INTERFACES.md](docs/INTERFACES.md) for CLI-first boundaries.

New capabilities must appear in the CLI before the terminal or native layers.
New narrative must react to event facts and cannot alter simulation state. New
physics must state equations, assumptions, units, validity, verification, and
validation status.

## Current development setup

Install the Go version declared in `go.mod`, then run:

```sh
go build ./...
go vet ./...
go test ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Run `gofmt` on every changed Go file. Total statement coverage must remain at or
above 80 percent.

## Pull requests

- Keep the change focused and explain the observable behavior.
- Add or update tests for behavior and failure cases.
- Update contracts and roadmap status only when implementation earns it.
- Preserve deterministic output and explicit seeds.
- Keep machine-readable stdout free of progress and diagnostics.
- Do not add generated binaries, coverage files, local paths, credentials,
  personal metadata, or raw automation receipts.
- Do not add prerecorded emissions as the source of simulated audio.
- Do not add browser shells, embedded webviews, or local web servers as desktop
  architecture.

For security-sensitive changes, include the attacker boundary, expected control,
and focused regression proof. Report existing vulnerabilities through the
private process in [SECURITY.md](SECURITY.md).

## Scientific honesty

Conservation does not prove a model is correct. Grid convergence does not prove
empirical validity. Alternate-law worlds can be internally consistent without
being validated descriptions of our universe. Documentation and certificates
must keep those claims separate.
