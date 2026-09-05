# Contributing

F.A.R.T. Lab is a silly idea taken seriously. Small, verified steps are better
than large speculative implementations.

Unless explicitly stated otherwise, intentionally submitted contributions are
accepted under the repository's [Apache License 2.0](LICENSE).

## Before proposing a change

Read:

- [README.md](README.md) for product identity and current status.
- [ROADMAP.md](ROADMAP.md) for sequencing and completion gates.
- [docs/PROJECT_LAYOUT.md](docs/PROJECT_LAYOUT.md) for source organization and
  dependency direction.
- [docs/RUST_CORE.md](docs/RUST_CORE.md) for the native core's supported subset,
  comparison policy, and dependency decisions.
- [docs/SIMULATION.md](docs/SIMULATION.md) for scientific claims.
- [docs/MODELS.md](docs/MODELS.md) for model registry, ensemble, and scientific
  machine-learning boundaries.
- [docs/GAMEPLAY.md](docs/GAMEPLAY.md) for procedural story rules.
- [docs/INTERFACES.md](docs/INTERFACES.md) for CLI-first boundaries.
- [docs/AUDIO.md](docs/AUDIO.md) for acoustics, Symphony, radio, and music assets.
- [docs/AGENT_PLAY.md](docs/AGENT_PLAY.md) for actions, observations, protocol
  adapters, fairness, and benchmark rules.
- [docs/AGENT_INTEGRATION.md](docs/AGENT_INTEGRATION.md) for the portable skill,
  executable recipes, and reviewed protocol baseline.
- [docs/CULTURE.md](docs/CULTURE.md) for cultural and public-interest safeguards.
- [docs/LOCALIZATION.md](docs/LOCALIZATION.md) for semantic, language, script,
  and optional communication-profile contracts.
- [docs/METROLOGY.md](docs/METROLOGY.md) for the Reference Pfft and traceability.
- [docs/SNOWFLAKES.md](docs/SNOWFLAKES.md) for record identity, optional
  context-occurrence identity, and artifacts.
- [docs/QUALITY.md](docs/QUALITY.md) for progressive engineering gates.
- [docs/CAPABILITY_REPORT.md](docs/CAPABILITY_REPORT.md) for evaluation
  disposition, owner-profile decomposition, unresolved semantics, and the
  report ratification gate.

New capabilities must appear in the CLI before the terminal or native layers.
New narrative must react to retained claims and cannot alter the authoritative
Lab account. New physics must state equations, assumptions, units, validity,
verification, and validation status.

## Current development setup

Install Go 1.27.1 as declared in `go.mod`. Documentation lint uses Node.js 26.8.1
from `.node-version` and npm 12.0.2 from `package.json`'s `packageManager` pin.
The native workspace pins Rust 1.98.1 in `rust-toolchain.toml`.
These are reviewed stable releases; use the
[toolchain policy](docs/QUALITY.md) when updating them. Then run:

```sh
npm ci --ignore-scripts
npm run lint:markdown
go mod verify
go run ./tools/repoquality repository
go build ./...
go test ./internal/registrationauthoritybinding -run '^TestRegistrationAuthorityBindingCorpus$'
go test ./internal/snapshotregistrationbinding -run '^TestSnapshotRegistrationBindingCorpus$'
go test ./internal/authoritymatching -run '^TestAuthorityMatchingCorpus$'
go test ./internal/authorityresolution -run '^TestAuthorityResolutionCorpus$'
go test ./internal/cataloglookup -run '^TestLookupCorpus$'
go test ./internal/catalogregistration -run '^TestRegistrationCorpus$'
go test ./internal/evaluation -run '^TestDispositionCorpus$'
go test ./internal/lawcatalog -run '^TestBuiltInCatalog$'
go test ./internal/cli -run '^TestLawCLITextAndJSONFixtures$'
go test ./internal/lawcatalog -run '^TestMinimalOpaqueContextHasNoLocalizedPresentationOrOptionalStructuralModule$'
go test ./internal/cli -run '^TestMinimalOpaqueLawInspectionJSONFixture$'
go test ./internal/scenarioprobe -run '^TestAtemporalProbeHasNoAmbientOrEarthRequirements$'
go test ./internal/cli -run '^TestScenarioCLITextAndJSONFixtures$'
go test ./internal/scenarioprobe -run '^TestMinimalOpaqueProbeRequiresNoLocalizedPresentationOrOptionalStructuralModule$'
go test ./internal/cli -run '^TestMinimalOpaqueScenarioJSONFixture$'
go test ./internal/scenarioprobe -run '^TestCaseOperationAbsenceIsInferredOnlyAfterSchemaValidation$'
go test ./internal/cli -run '^TestMultiLawProbeLimitDoesNotInferCompatibility$'
go test ./internal/scenarioprobe -run '^TestMinimalOpaqueUnresolvedCapabilityStopsAtOuterEnvelope$'
go test ./internal/cli -run '^TestMinimalOpaqueUnresolvedCapabilityCLIContract$'
go test ./internal/idealmixturereservoir -run '^TestSyntheticMixtureClosedForms$'
go test ./internal/reservoirprediction -run '^TestPredictSyntheticClosedForms$'
go test ./internal/cli -run '^TestReservoirCLITextAndJSONFixtures$'
go test ./internal/restrictionflow -run '^TestChokedGammaFifteenClosedForm$'
go test ./internal/restrictionprediction -run '^TestPredictChokedClosedForm$'
go test ./internal/restrictionhistory -run '^TestConstantChokedHistoryClosedForm$'
go test ./internal/restrictionhistoryprediction -run '^TestPredictConstantChokedHistory$'
go test ./internal/cli -run '^TestRestrictionCLITextAndJSONFixtures$'
go test ./internal/coupledblowdown -run '^TestAdiabaticPathMatchesReservoirEndpoint$'
go test ./internal/walkcase -run '^TestWalkOperationsOnIsothermalFixture$'
go test ./internal/cli -run '^TestWalkCLISimulateAndFailures$'
go test ./internal/cli -run '^TestHelpRoutes$'
go vet ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.7.0 ./...
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
go test ./...
go test -shuffle=on -count=20 ./...
go run ./tools/repoquality fuzz --time 5s
mkdir -p artifacts/coverage
go test '-coverprofile=artifacts/coverage/go.out' ./...
go run ./tools/repoquality coverage --profile artifacts/coverage/go.out --aggregate 90 --package 80
```

Run `gofmt` on every changed Go file. Total statement coverage must remain at or
above 90 percent, and every non-generated package must remain at or above the 80
percent floor. Domain, solver, archive, replay, and protocol code receive stricter
changed-line and mutation gates described in [docs/QUALITY.md](docs/QUALITY.md).

For native changes, also run:

```sh
cargo fmt --all --check
cargo clippy --locked --workspace --all-targets --all-features -- -D warnings
cargo doc --locked --workspace --no-deps
cargo test --locked --workspace --all-features
cargo build --locked --release --workspace
cargo deny --locked check
cargo llvm-cov --locked --workspace --all-features --json --summary-only --output-path artifacts/coverage/rust.json
go run ./tools/repoquality rust-coverage --profile artifacts/coverage/rust.json --aggregate 90 --package 80
```

Install `cargo-deny` 0.20.2 and `cargo-llvm-cov` 0.9.0 with
`cargo install --locked --version <version> <tool>`. Set `RUSTDOCFLAGS=-D warnings`
for the documentation gate. Set `FARTAPP_GO_ORACLE` to the absolute path of a
freshly built Go executable when running cross-language tests. CI requires that
comparison on each supported system. See [QUALITY.md](docs/QUALITY.md) for the
distinction between current checks and broader production gates.

Repository-policy checks are the Go `repoquality` command documented in
[docs/QUALITY.md](docs/QUALITY.md). Do not add new policy to the optional
PowerShell wrappers or introduce another task runner. Wrappers may only forward
arguments and exit status to the Go command.

## Repository structure and review discipline

Keep source in `cmd/`, `internal/`, and `crates/` as described in the
[layout guide](docs/PROJECT_LAYOUT.md). Root files are entry-point documents,
policies, and conventional tool manifests. Put generated local outputs under
ignored `artifacts/` or their tool's standard output directory.

For a new package or a meaningful file split, describe the responsibility of
each file and the permitted dependency direction in the pull request. Prefer
one coherent responsibility per file. Do not combine wire decoding, domain
validation, numerical stepping, report construction, and interface presentation
in one implementation unit, and do not fragment a small responsibility merely
to reduce line counts.

Before requesting review:

- Remove dead fields, unused helpers, redundant wrappers, unreachable branches,
  speculative interfaces, and comments that repeat the code.
- Reuse parsing or validation only when the complete contract is identical.
  Similar syntax with different authority, bounds, or error semantics is not a
  shared abstraction.
- Give each behavior test an independent expected result or a property that a
  plausible mutation can violate. Coverage without discriminating assertions
  is insufficient.
- Exercise invalid, forged, boundary, and mismatch cases at the layer that owns
  the invariant.
- Review all assisted changes line by line for invented APIs, fake citations,
  dependency drift, hidden platform assumptions, and claims not earned by the
  implementation.

The same acceptance standard applies to every contribution regardless of the
tools used to prepare it. Do not add assistant names, generator bylines,
authorship labels, or attribution comments to source, documentation, commits,
fixtures, or product output.

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
- Do not accept a change merely because aggregate coverage increased or all
  existing tests remained green. Explain which defect the new evidence can
  detect.

For security-sensitive changes, include the attacker boundary, expected control,
and focused regression proof. Report existing vulnerabilities through the
private process in [SECURITY.md](SECURITY.md).

## Scientific honesty

Conservation does not prove a model is correct. Grid convergence does not prove
empirical validity. Alternate-law worlds can be internally consistent without
being validated descriptions of our universe. Documentation and certificates
must keep those claims separate.
