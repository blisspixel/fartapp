# Interface and release strategy

F.A.R.T. Lab is CLI first by design. The command line is the complete,
scriptable scientific instrument and a finished way to play. Later interfaces
improve observation and sensation without becoming alternate sources of physics,
story truth, or hidden capability.

## Permanent promotion rule

Every physical model, law pack, storylet predicate, mode, challenge primitive,
and proof tool ships through the same gates:

1. Implement and verify it in the headless core and CLI.
2. Instrument and control it in the cross-platform terminal UI.
3. Present it in the native desktop app.

If a native effect cannot be reproduced, inspected, exported, or explained from
a CLI archive, it is not ready to ship.

## Architecture

The planned Rust workspace has strict dependency direction:

```text
fart-domain      quantities, laws, schemas, regimes, scenario types
fart-core        deterministic solvers and proof ledgers
fart-archive     .fart read/write, validation, hashes, migrations
fart-audio       deterministic offline procedural synthesis
fart-narrative   seeded content, fact predicates, and episode timelines
fart-services    simulate, play, inspect, sweep, translate, verify, export
fart-cli         command presentation
fart-tui         terminal presentation over fart-services
fart-gdext       thin native Godot adapter
xtask            fixtures, schemas, completions, manuals, release tasks
```

The current Go executable remains a tiny reference oracle and fixture generator.
It is not a runtime dependency of the production application. Rust must match
versioned Go fixtures within documented tolerances before replacing an oracle
model.

The desktop application uses Godot with a native Rust GDExtension. The core has
no dependency on Godot. The render thread consumes immutable or double-buffered
snapshots from a background simulation worker. There is no browser runtime,
webview, HTML UI, local HTTP service, Electron, or Tauri shell.

## Interface 1: CLI Lab

The CLI stands on its own for play, science, automation, regression, and people
who prefer terminals.

### Immediate play

```console
fart
fart quick --seed 42 --record run.fart
fart broadcast --seed 42 --length standard
fart freestyle bathroom.toml --set reservoir.pressure="106 kPa"
```

When `fart` runs in an interactive terminal, it launches Quick Play. With
redirected input or output, it prints concise help and exits without prompting.
No play command writes into the current directory unless the player explicitly
requests `--record` or `--output`.

Quick Play is a polished result, not a random parameter dump. Broadcast resolves
a world, source, story, scenario, event, and narrative timeline deterministically.
Freestyle is guided when interactive, but every choice has a scenario or flag
equivalent.

### Laboratory commands

```console
fart scenario init -o bathroom.toml
fart scenario validate bathroom.toml
fart simulate bathroom.toml -o run.fart
fart inspect run.fart --at "1.2 s"
fart explain run.fart --why regime.choked
fart provenance run.fart --to consumers/audio
fart branch run.fart --set exterior.pressure="0 Pa" -o vacuum.fart
fart sweep bathroom.toml --vary reservoir.pressure="105 kPa..800 kPa" --steps 64
fart compare small.fart large.fart --nondimensional
fart translate source.fart --target-world hush-3.toml --mode strict
fart verify run.fart --refine timestep
fart replay run.fart --mode exact
fart export run.fart --format wav -o emission.wav
fart lab run.fart
fart schema scenario
fart completions powershell
```

Command names stay provisional until the typed command and schema RFC is
accepted. Capabilities and output discipline are not provisional.

### Global output contract

Planned global controls include:

```text
--format human|json|jsonl|csv
--color auto|always|never
--unicode auto|always|never
--progress auto|always|never
--quiet
--log-level error|warn|info|debug|trace
--threads N
--seed VALUE
```

Rules:

- Stdout contains only the requested result.
- Diagnostics and interactive progress use stderr.
- JSON contains no ANSI sequences and validates against a versioned schema.
- JSONL emits typed progress records and one terminal result record.
- Progress defaults off when stderr is not a terminal.
- Broken pipes end quietly.
- Numeric parsing is locale-independent and units are explicit.
- `-` represents stdin or an explicitly requested artifact stream, never a mix
  of binary output and diagnostics.
- Help includes real examples, units, validity limits, exit codes, and recovery
  suggestions.
- Invalid physical states fail with exact field paths and are not silently
  clamped into a different scenario.
- Convenience controls expose the explicit scenario diff they produce.
- Every displayed scientific value can reveal its source, canonical unit,
  uncertainty, validity domain, derivation, and provenance path on demand.

Stable exit-code families should distinguish success, command or input error,
unsupported law or regime, numerical failure, verification failure, archive or
filesystem failure, replay mismatch, and cancellation.

### Identity contract

The interface never collapses these identities:

| Identity | Changes when | Stability promise |
| --- | --- | --- |
| Scenario | Normalized laws, inputs, or seed change | Unit spelling and field order do not change it |
| Physical result | Solver, numerics, or authoritative state changes | Presentation cannot change it |
| Narrative | Resolved world, storylets, or narrative streams change | Camera and terminal width cannot change it |
| Presentation | Language, layout, device, camera, or accessibility changes | No physical claim follows from it |
| Archive bytes | Serialization or container bytes change | Byte equality is never confused with semantic equality |

`fart provenance` traverses the typed event graph. `fart explain` presents a
causal path and a runnable counterfactual. `fart branch` creates a new scenario
with explicit ancestry and never mutates a certified source.

### CLI experience bar

- Startup and first meaningful output feel immediate on every supported system.
- Human hierarchy, spacing, labels, plots, and tables adapt to terminal width.
- Plain output is stable for pipes, snapshots, logs, and screen readers.
- Color, box drawing, animation, hyperlinks, audio, and Unicode are enhancements,
  never the only information channel.
- `NO_COLOR`, `TERM=dumb`, redirected streams, and no-audio systems work cleanly.
- Errors identify the bad value, expected domain, model assumption, and most
  useful next action.
- Commands have predictable nouns, verbs, defaults, `--dry-run`, and `--force`
  semantics where applicable.
- Shell completions, manual pages, schemas, examples, and machine-output
  contracts are generated from the same typed command model.
- Golden tests cover interactive and piped output, documented widths, ASCII and
  Unicode, color modes, hostile strings, and Windows, macOS, and Linux paths.
- No command needs a display server, browser, account, or network access.

Release candidates publish p50 and p95 measurements on named Windows, macOS,
and Linux reference systems. Initial budgets are less than 250 ms to `--help`,
less than 250 ms to the Quick Play title and seed, less than 1 s for an ordinary
analytical event, less than 100 ms to acknowledge cancellation, and less than
1 s to reach a safe cancellation boundary outside a documented atomic section.
The benchmark RFC also sets resident-memory and ordinary-archive-size budgets
after the first measured implementation. Regressions require evidence, not a
quietly weakened target.

### Cancellation and atomic output

The first interrupt requests cooperative cancellation. A second terminates
immediately. Long-running solver, refinement, audio, and archive loops check a
cancellation token at bounded intervals.

Archives are built in a same-directory temporary file. The writer validates,
flushes, synchronizes, closes, and atomically renames the final archive, with a
best-effort parent-directory sync. Existing targets require `--force`.
Cancellation leaves either no final archive or an explicitly requested,
uncertified `.partial.fart`.

## Event and episode archives

The event archive is the boundary between computation and presentation. A
documented `.fart` package may contain:

```text
mimetype
manifest.json
scenario.json
history.jsonl
certificate.json
fields/<content-hash>.bin
consumers/audio.json
```

Canonical JSON metadata uses RFC 8785. Schemas use JSON Schema 2020-12. Large
typed histories begin with the simplest bounded, streamable representation that
meets measured budgets. Apache Arrow IPC or another columnar layer is added
behind an isolated adapter only when profiling proves it necessary. Every
member has a SHA-256 content hash, and the manifest commits to sorted names and
hashes. Hash logical content rather than incidental ZIP compression bytes.

Writers use fixed member order, normalized timestamps, and no machine-specific
metadata. Migration creates a new archive with provenance and never mutates a
certified source. Readers reject duplicate names, traversal, links, oversized
members, decompression bombs, impossible array lengths, nonfinite invalid JSON,
and hash mismatches before allocating or parsing payloads.

Initial third-party content packs are versioned, declarative data only. They
cannot load native code, access the network, name arbitrary local files, or
bypass law-capability and resource checks.

A `.fartshow` episode bundle adds resolved world, culture, narrative streams,
fact-provenance timeline, transcript, and optional presentation assets. Seeds
alone are not sufficient for archival replay because content packs evolve.

## Procedural audio

The normative offline export begins with deterministic 48 kHz PCM16 RIFF/WAVE.
The pure audio core emits sample buffers from event source terms, a documented
stochastic closure, explicit stream keys, exact rational sample clock, room
model, limiter, quantization, and dither policy.

The TUI may offer optional audio-device playback. Offline WAV must work without
an audio device. Godot consumes the same synthesized frames in real time. Device
resampling is presentation behavior and does not redefine the certified WAV.

## Interface 2: Terminal Lab

The TUI should feel like htop for an absurd research instrument. It ships in the
single `fart` binary as `fart lab`, while remaining a separate crate over the
same services.

Views include overview, emitter, interface, plume and payload, acoustics, proof
ledger, timeline, translation, and Broadcast narration.

Responsive breakpoints:

- `160x48`: wide plots, ledger, timeline, and narrative.
- `120x36`: standard multi-pane laboratory.
- `80x24`: tabbed compact mode.
- `60x18`: minimum gauges and alerts.
- Smaller: clear size guidance and a plain-mode command.

Quality requirements:

- Keyboard-only operation with remappable keys.
- Mouse support is optional.
- ASCII, 16-color, 256-color, and truecolor modes.
- Grapheme-aware truncation and terminal-cell width calculation.
- Terminal state restoration after success, error, panic, interrupt, and resize.
- Append-only `--plain --watch` output for screen readers, logs, and weak
  terminals.
- Audio never autoplays.
- Snapshot tests target the cell buffer, not terminal escape sequences.
- PTY and ConPTY smoke tests cover launch, resize, quit, cancellation, panic,
  and restoration.
- Every action can display or copy its equivalent CLI command.

## Interface 3: Native Lab

The native Godot application ships as a normal Windows, macOS, and Linux app.
It adds real-time 3D presentation, spatial and procedural audio, haptics, rooms,
particles, deposition, and consequence systems while consuming the same event
identity and service layer.

Native quality includes keyboard, mouse, controller, remapping, accessibility
setup before first play, safe saves, crash handling, clean uninstall, native
packaging, signing, notarization where applicable, and deterministic archive
export that the CLI can reproduce.

The first native room is the Tiled Chamber. Before native implementation begins,
the same scenario must already run, verify, replay, translate, render audio, and
present in CLI and TUI. The native milestone proves that every audiovisual and
haptic consumer cites the same archive-compatible event history.

## Cross-platform release gate

Every merge eventually enforces:

- Go formatting, vet, tests, race checks where supported, and at least 80 percent
  statement coverage.
- Rust formatting, Clippy with warnings denied, tests, documentation, oracle
  parity, and at least 80 percent coverage for solver and archive crates.
- Windows, macOS, and Linux replay tolerances.
- Dependency advisory, license, source, and duplicate-version policy.
- Schema validation, archive fuzzing, cancellation and atomic-write fault tests,
  secret scanning, and generated-artifact cleanliness.
- TUI snapshots and terminal restoration smoke tests.
- Headless Godot extension smoke tests once native work begins.

Release archives include executables, licenses, README, completions, manual
pages, schemas, checksums, SBOM, and provenance. macOS uses Developer ID,
hardened runtime, notarization, and stapling. Windows uses Authenticode with a
trusted timestamp. Linux support names its glibc baseline and package format
instead of claiming one build runs everywhere.

The native-start gate requires stable public CLI and TUI releases, a certified
Tiled Chamber archive, procedural WAV export, cross-platform QA, and a
GDExtension parity test.

## Product identity

The CLI makes the science and comedy complete. The Terminal Lab makes the event
legible in motion. The native app makes it visceral. None is a disposable
wrapper around another.

The research sources are listed in [RESEARCH.md](RESEARCH.md).
