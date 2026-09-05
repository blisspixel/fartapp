# SCOPE

## Current implementation

The repository preserves the original intensity CLI in Go and Rust. Go also
implements strict scenario and law inspection, bounded analytical reservoir and
restriction calculations, coupled blowdown and refinement, and provisional
retained-evidence operations. Rust implements the stateless reservoir subset.
The [README](../README.md) records available commands and their limits; the
full production service and certified archive milestones remain open.

## Product scope

In scope, progressively:

- A complete cross-platform CLI for Quick Play, seeded Broadcast, Freestyle,
  challenges, simulation, translation, proof, replay, audio, Symphony, radio,
  agent play, and export.
- Source-neutral law, world, emitter, interface, payload, observer, numerical,
  event, episode, and certificate contracts.
- A deterministic Rust production core conforming to Go reference fixtures.
- An htop-style Terminal Lab over the same services.
- A native Godot app over the same archive and core, never a web wrapper.
- Optional MCP and A2A adapters, human and agent benchmark tracks, and opt-in
  native automation over one canonical play service.
- Earth continuum physics first, then separately verified specialist and
  fictional-law profiles.

Out of scope until their roadmap gates are met:

- Claims of medical, biological, aerospace, pressure-vessel, or safety-grade use.
- Native-only physics, story facts, challenge variables, or hidden controls.
- Adapter-owned gameplay logic, hidden agent observations, compute-priced rank,
  or protocol access to shell, credentials, or arbitrary files and networks.
- Treating a 2D visualization as validated 3D physics or another universe.
- Feeding planetary, stellar, plasma, or relativistic values into a bathroom
  model.
- Calling a fictional-law profile empirically validated.

## Evidence

- [Roadmap](../ROADMAP.md)
- [Simulation contract](../docs/SIMULATION.md)
- [Gameplay contract](../docs/GAMEPLAY.md)
