# MODEL

- **Implemented language:** Go
- **Planned production languages:** Rust and Godot-supported native scripting
- **Updated:** 2026-09-01

## Understanding

The current product is a five-level Go CLI and test suite. The ratified direction
is a CLI-first simulation and procedural comedy game, followed by an htop-style
Terminal Lab and a native Godot application. The current docs are design
contracts, not implemented physics.

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
fart.go                   current reference behavior
fart_test.go              current tests
main.go                   current CLI entry
go.mod                    current Go module
```
