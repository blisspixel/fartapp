# Contributing

F.A.R.T. Lab is a silly idea taken seriously. Small, verified steps are better
than large speculative implementations.

Unless explicitly stated otherwise, intentionally submitted contributions are
accepted under the repository's [Apache License 2.0](LICENSE).

## Before proposing a change

Read:

- [README.md](README.md) for product identity and current status.
- [ROADMAP.md](ROADMAP.md) for sequencing and completion gates.
- [docs/SIMULATION.md](docs/SIMULATION.md) for scientific claims.
- [docs/GAMEPLAY.md](docs/GAMEPLAY.md) for procedural story rules.
- [docs/INTERFACES.md](docs/INTERFACES.md) for CLI-first boundaries.
- [docs/AUDIO.md](docs/AUDIO.md) for acoustics, Symphony, radio, and music assets.
- [docs/AGENT_PLAY.md](docs/AGENT_PLAY.md) for actions, observations, protocol
  adapters, fairness, and benchmark rules.
- [docs/CULTURE.md](docs/CULTURE.md) for cultural and public-interest safeguards.
- [docs/LOCALIZATION.md](docs/LOCALIZATION.md) for semantic, language, script,
  and optional communication-profile contracts.
- [docs/METROLOGY.md](docs/METROLOGY.md) for the Reference Pfft and traceability.
- [docs/SNOWFLAKES.md](docs/SNOWFLAKES.md) for record identity, optional
  context-occurrence identity, and artifacts.
- [docs/QUALITY.md](docs/QUALITY.md) for progressive engineering gates.

New capabilities must appear in the CLI before the terminal or native layers.
New narrative must react to retained claims and cannot alter the authoritative
Lab account. New physics must state equations, assumptions, units, validity,
verification, and validation status.

## Current development setup

Install the Go version declared in `go.mod`. Documentation lint uses Node.js 24
and the locked npm toolchain. Then run:

```sh
npm ci --ignore-scripts
npm run lint:markdown
pwsh ./scripts/check-dependencies.ps1
pwsh ./scripts/check-links.ps1
pwsh ./scripts/check-media.ps1
go mod verify
go build ./...
go test ./internal/lawcatalog -run '^TestBuiltInCatalog$'
go test . -run '^TestLawCLITextAndJSONFixtures$'
go test ./internal/lawcatalog -run '^TestMinimalOpaqueContextHasNoLocalizedPresentationOrOptionalStructuralModule$'
go test . -run '^TestMinimalOpaqueLawInspectionJSONFixture$'
go test ./internal/scenarioprobe -run '^TestAtemporalProbeHasNoAmbientOrEarthRequirements$'
go test . -run '^TestScenarioCLITextAndJSONFixtures$'
go test ./internal/scenarioprobe -run '^TestMinimalOpaqueProbeRequiresNoLocalizedPresentationOrOptionalStructuralModule$'
go test . -run '^TestMinimalOpaqueScenarioJSONFixture$'
go test ./internal/scenarioprobe -run '^TestCaseOperationAbsenceIsInferredOnlyAfterSchemaValidation$'
go test . -run '^TestMultiLawProbeLimitDoesNotInferCompatibility$'
go test . -run '^TestHelpRoutes$'
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
go test ./...
go test -shuffle=on -count=20 ./...
go test -run=^$ -fuzz=FuzzRun -fuzztime=5s .
go test -run=^$ -fuzz=FuzzValidate -fuzztime=5s ./internal/scenarioprobe
go test '-coverprofile=coverage.out' ./...
pwsh ./scripts/check-go-coverage.ps1 -ProfilePath coverage.out -AggregateMinimum 90 -PackageMinimum 80
```

Run `gofmt` on every changed Go file. Total statement coverage must remain at or
above 90 percent, and every non-generated package must remain at or above the 80
percent floor. Domain, solver, archive, replay, and protocol code receive stricter
changed-line and mutation gates described in [docs/QUALITY.md](docs/QUALITY.md).

## Pull requests

- Keep the change focused and explain the observable behavior.
- Add or update tests for behavior and failure cases.
- Update contracts and roadmap status only when implementation earns it.
- Preserve deterministic output and explicit seeds where generation applies.
- Keep machine-readable stdout free of progress and diagnostics.
- Do not add generated binaries, coverage files, local paths, credentials,
  personal metadata, or raw automation receipts.
- Do not add prerecorded emissions as the source of simulated audio.
- Do not add generated or third-party media without its compact asset manifest,
  rights approval, source review, and content review.
- Do not add secrets or paid generation calls to pull-request or ordinary CI
  workflows. Development tools read credentials only from ignored environment
  files or the process environment.
- Do not add browser shells, embedded webviews, or local web servers as desktop
  architecture.
- Do not add gameplay logic, hidden observations, or solver mutation access to
  CLI, MCP, A2A, TUI, Godot, spectator, or automation adapters.

For security-sensitive changes, include the attacker boundary, expected control,
and focused regression proof. Report existing vulnerabilities through the
private process in [SECURITY.md](SECURITY.md).

## Scientific honesty

Conservation does not prove a model is correct. Grid convergence does not prove
empirical validity. Alternate-law worlds can be internally consistent without
being validated descriptions of our universe. Documentation and certificates
must keep those claims separate.
