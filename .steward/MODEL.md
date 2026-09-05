# MODEL

- **Implemented languages:** Go and Rust
- **Planned native presentation:** Godot-supported native scripting
- **Updated:** 2026-09-05

## Understanding

The current product preserves the five-level toy and adds Go law inspection,
scenario validation, analytical reservoir and restriction oracles, coupled
blowdown, explicit refinement, and a provisional retained-evidence carrier.
The four-crate Rust foundation implements the toy and stateless reservoir
subset. The [README](../README.md) distinguishes working features from planned
ones; the [roadmap](../ROADMAP.md) owns milestone gates.

The direction remains a CLI-first simulation and procedural comedy game,
followed by an htop-style Terminal Lab and a native Godot application.
Executable analytical checks establish bounded software evidence; broader
design contracts, empirical validation, and certified archives remain open.

## Structure

```text
README.md                 product overview and current CLI
ROADMAP.md                authoritative progressive plan
docs/SIMULATION.md        scientific and numerical contract
docs/GAMEPLAY.md          modes, story, seeds, and progression
docs/INTERFACES.md        CLI, TUI, native, archive, and release contract
docs/AUDIO.md             acoustics, sonification, Symphony, and radio contract
docs/AGENT_PLAY.md        play service, agent protocols, grades, and benchmarks
docs/CULTURE.md           cultural, religious, economic, and review safeguards
docs/RESEARCH.md          authoritative source basis
cmd/fartapp/              Go process entry point
internal/cli/             Go command behavior and CLI tests
internal/                 private Go domains, oracles, adapters, and checks
crates/fart-*/            native Rust domain, core, services, and CLI
testdata/                 shared authored inputs and conformance fixtures
artifacts/                ignored local outputs
go.mod                    Go module and compiler pin
Cargo.toml                Rust workspace and dependency policy
rust-toolchain.toml       Rust compiler and component pin
```

See [project layout](../docs/PROJECT_LAYOUT.md) for ownership and dependency
direction. Files under `.steward/DECISIONS/` retain historical decisions and
their original command paths.
