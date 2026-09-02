# Interface and release strategy

F.A.R.T. Lab is CLI first by design. The planned 1.0 command line is specified
as the complete, scriptable scientific instrument and a finished way to play.
Later interfaces improve observation and sensation without becoming alternate
sources of physics, story truth, or hidden capability.

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
fart-compute     CPU and accelerator-neutral field-kernel contract
fart-field       optional C++20 Kokkos field solver behind a narrow C ABI
fart-metal       optional non-certifying native Apple Metal preview backend
fart-archive     .fart read/write, validation, hashes, migrations
fart-audio       deterministic offline procedural synthesis
fart-score       diagnostic sonification and Symphony score mappings
fart-radio       station catalogs, schedules, captions, and presentation state
fart-narrative   seeded content, fact predicates, and episode timelines
fart-services    canonical play reducer, observations, simulation, and proof
fart-agent-protocol  actions, observations, roles, budgets, and receipts
fart-cli         command presentation
fart-tui         terminal presentation over fart-services
fart-mcp         optional MCP adapter over fart-services
fart-a2a         optional A2A adapter over fart-services
fart-gdext       thin native Godot adapter
xtask            fixtures, schemas, completions, manuals, release tasks
```

The current Go executable remains a tiny reference oracle and fixture generator.
It is not a runtime dependency of the production application. Rust must match
versioned Go fixtures within documented tolerances before replacing an oracle
model.

### Language policy

Languages are selected for the strongest long-term implementation boundary, not
the shortest prototype:

- **Go:** the small, independently auditable analytical oracle already present.
  It owns fixtures, not the production runtime.
- **Rust:** domain types, units, law schemas, solvers, the CPU reference backend,
  provenance, archives, services, CLI, TUI, MCP, A2A, deterministic DSP,
  tooling, and native boundaries. Pure domain and solver crates forbid unsafe
  code.
- **C++20 with Kokkos:** optional production field kernels built separately for
  CPU, CUDA, HIP, and SYCL behind one narrow versioned C ABI. They do not own
  domain rules, schemas, archives, or certificate policy.
- **Metal Shading Language:** optional Apple preview kernels plus native host
  glue, with no FP64 quantitative CFD claim.
- **Godot scene and shader languages:** native presentation, rendering, and GPU
  effects only. They cannot define physics, scoring, or archive truth.
- **Declarative data:** scenarios, law packs, storylets, stations, localization,
  schemas, and manifests under bounded parsers.

Python, JavaScript, JVM, browser, and web runtimes are not application
dependencies. Development-only analysis may use a separately locked tool when
it produces a reviewed fixture or artifact that the shipped product can verify
without that tool.

C, C++, or Fortran otherwise enters only for a measured specialist numerical
library that cannot reasonably be implemented or matched in safe Rust. Every
native boundary uses a minimal C ABI, pinned source, license and vulnerability
review, explicit memory and thread ownership, sanitizer and fuzz coverage,
independent reference cases, and a safe replacement path. Convenience is not
enough to justify FFI.

Mojo remains an evaluated research candidate. Vulkan and WebGPU compute remain
presentation candidates rather than quantitative CFD backends. Promotion
depends on accuracy, toolchain stability, native platform coverage, maintenance
cost, and measured performance against the project suite. The complete policy
is in [COMPUTE.md](COMPUTE.md).

### Dependency budget

Standard libraries and small owned modules are preferred, but “few dependencies”
does not justify homemade cryptography, Unicode, archive security, or numerical
algorithms whose established implementation is safer. Every production
dependency records:

- Exact purpose and why existing code cannot satisfy it safely.
- Direct and transitive license, source, maintenance, and vulnerability status.
- Features enabled, code-size and startup effect, unsafe surface, build scripts,
  network behavior, and platform support.
- Owning crate, upgrade policy, exit strategy, and last review.

Default features are disabled unless reviewed. Lockfiles are committed. Duplicate
versions, git dependencies, build scripts, native code, macros, and unsafe code
receive explicit policy. Initial law and content packs are data, not dynamic
libraries. Release builds can be produced from a reviewed source closure without
fetching mutable code during compilation.

The desktop application uses Godot with a native Rust GDExtension. The core has
no dependency on Godot. The render thread consumes immutable or double-buffered
snapshots from a background simulation worker. There is no browser runtime,
webview, HTML UI, local HTTP service, Electron, or Tauri shell.

The canonical `PlayService` owns session state. Its public operations start,
observe, list actions, act, checkpoint, branch, finish, and export. CLI, TUI,
Godot, MCP, A2A, spectators, and native automation are adapters over that
service. They cannot import solver mutation APIs or maintain private gameplay
branches. The service can remain in process; an explicitly started A2A endpoint
is interoperability infrastructure, not the application UI.

A canonical `CapabilityService` returns the same typed `CapabilityReport` to
CLI, TUI, native, MCP, and A2A adapters. It separates law-defined concepts,
implementation and closure availability, scenario applicability, evidence
grade, trust policy, backend feasibility, and resource refusal. Adapters may
filter or present the report, but cannot invent capability from an absent pane,
device feature, or protocol method.

## Interface 1: CLI Lab

The CLI is designed to stand on its own for play, science, automation,
regression, and people who prefer terminals.

Commands written as `fart` in this section describe the planned installed 1.0
product unless a subsection explicitly says they exist now. The current Go
oracle is named `fartapp`; it implements the permanent intensity path, law
catalog inspection, scenario-document validation, and their exact help routes.

### Immediate play

```console
fart
fart quick --seed 42 --record run.fart
fart broadcast --seed 42 --length standard
fart ask "a tiny dry pfft in low gravity, under 20 J" --dry-run
fart freestyle reference-enclosure.toml --set emitter.pressure="106 kPa"
fart play start challenge:dry-c-sharp-01 --seed 42 --json
fart play act PLAY_HANDLE --action set_pressure --value "180 kPa" --json
```

Planned 1.0 behavior: when `fart` runs in an interactive terminal, it launches
Quick Play. With
redirected input or output, it prints concise help and exits without prompting.
No play command writes into the current directory unless the player explicitly
requests `--record` or `--output`.

Quick Play is a polished result, not a random parameter dump. Broadcast resolves
a law context, scope, and realization deterministically, then adds a world,
source, story, observation, or narrative ordering only when supported.
Freestyle is guided when interactive, but every choice has a scenario or flag
equivalent.

### Natural-language instrument

Natural language is a first-class authoring surface for humans and agents, not a
second simulation engine. A request compiles into one reviewable typed proposal:

Display language never enters an event identifier, field name, regime code, or
archive comparison. Locale packs and optional communication profiles follow
[LOCALIZATION.md](LOCALIZATION.md); the complete typed machine surface remains
usable without any natural-language model.

```console
fart ask "one polite dry pfft in a low-gravity station, under 20 J" --dry-run
fart ask request.txt --accept --record station.fart
fart explain-plan station.fart --from-language
```

The compiler returns:

- Applicable typed law contexts, scope, structures, capabilities, measurement
  interactions, view profiles, content limits, and extension fields. Quantities,
  units, source, exterior, and observer appear only when supported.
- Assumptions, ambiguities, unsupported requests, and confidence by field.
- A canonical scenario diff and action plan.
- Applicable compute, action, energy, and recording budgets.
- A stable interpretation receipt linking source spans to typed fields.

`--dry-run` is the default whenever wording is ambiguous, hazardous, expensive,
or would cross a content boundary. Acceptance freezes the typed proposal before
the occurrence. Later changes create an explicit diff and new scenario identity.

The ordinary parser and curated request grammar work fully offline. Optional
local or remote language-model adapters may propose typed fields, but they are
never required, never receive solver mutation access, and never bypass schema,
unit, capability, safety, content, or budget validation. Provider text is
untrusted input. The accepted typed proposal, not a model transcript, controls
the event.

MCP exposes the same draft, validate, accept, and explain operations. Agents can
use natural language for expressive play while exact JSON remains available for
benchmarking and automation. Same accepted proposal plus same declared event
inputs produces the same numerical reconstruction regardless of whether it was
authored through prose, flags, JSONL, MCP, TUI, or native controls.

### Command grammar

The command tree optimizes for recognition and learning. High-frequency play
modes stay short:

```console
fart
fart quick
fart broadcast
fart ask "a very small dry event in a 90 000 m3 station"
fart freestyle reference-enclosure.toml
fart lab
```

Scientific resources use consistent noun-then-verb groups:

```console
fart reference inspect
fart reference realize measured-enclosure.toml
fart scenario init
fart scenario validate reference-enclosure.toml
fart scenario diff first.toml second.toml
fart event run reference-enclosure.toml -o run.fart
fart event inspect run.fart
fart event explain run.fart --why regime.choked
fart event verify run.fart --refine timestep
fart event reconstruct run.fart
fart trace replay run.fart
fart artifact plumeprint run.fart
fart artifact grow run.fart
fart audio render run.fart
fart symphony render run.fart
fart radio play drift-93-7
fart play start challenge:dry-c-sharp-01
```

The same verbs keep the same meaning across nouns. `inspect` is read-only,
`validate` checks without mutation, `run` creates an occurrence, `reconstruct`
creates a new verification operation, `replay` presents retained evidence, and
`export` creates a representation. There is no second `upgrade` command that
competes with `update`.

The root command never treats an unknown word as a natural-language request or
catch-all action. A typo prints likely intended commands but does not execute
one. `-h`, `--help`, `fart help`, `fart help event run`, and
`fart event run --help` all work.

### Helpers and progressive disclosure

```console
fart help event run
fart examples reference
fart examples --search "vacuum"
fart doctor
fart doctor --json
fart version
fart version --json
fart config list
fart config get output.color
fart config set output.color never
fart config path
fart config validate
fart schema scenario --version 1
fart completions powershell
fart man event-run
fart explain-error FART-E-PHYS-0042
```

Help begins with a clear purpose and usage, then includes applicable arguments,
units, defaults, validity, output, working examples, exit codes, recovery, and
related commands. Errors contain a
stable code, the bad field or token, why it failed, what remains unchanged, and
the most useful recovery command. `doctor` checks installation provenance,
configuration, directories, terminal features, audio capability, plugin and
content manifests, update trust root, and platform prerequisites without
uploading anything.

Configuration precedence is explicit:

```text
compiled safe default
  < system configuration
  < user configuration
  < project configuration when explicitly trusted
  < environment variable
  < command-line flag
```

`fart config explain KEY` shows the winning value and every overridden source.
Unknown keys and insecure permissions are errors or prominent diagnostics, not
silent omissions. Secrets never belong in ordinary configuration or diagnostic
output.

### Secure update family

`fart update` is an explicit, user-initiated network operation. The application
does not silently phone home. Its family is:

```console
fart update check
fart update
fart update --version 1.2.3
fart update --channel stable
fart update --dry-run
fart update --yes
fart update rollback
fart update status --json
```

An installation records how it was installed. Package-managed copies do not
overwrite themselves. `fart update` prints the exact Homebrew, WinGet, Linux
package, or other supported manager command and exits with a documented status.
A standalone archive installation can update itself.

The standalone updater uses a TUF-style metadata hierarchy as its release trust
architecture:

1. Loads embedded root metadata with offline threshold keys and a monotonically
   versioned root-rotation policy.
2. Fetches and verifies expiring timestamp, snapshot, targets, and delegated
   channel and platform metadata using consistent snapshots.
3. Rejects rollback, freeze, mix-and-match, wrong-platform, wrong-architecture,
   and version-confusion responses.
4. Downloads to a same-filesystem staging location with strict size limits.
5. Verifies metadata, length, SHA-256 digest, target path, release identity, and
   platform signing where applicable before execution. A Sigstore transparency
   bundle may be required as target provenance, but never replaces the TUF role
   and threshold checks.
6. Runs a staged `--version --json` and self-check.
7. Atomically activates the new executable or schedules replacement through a
   minimal signed helper on Windows.
8. Preserves one verified rollback version and a non-secret receipt.
9. Restores the previous version automatically if activation health fails.

The updater never requests elevation, edits an unrelated directory, follows an
untrusted redirect to a new origin, executes installer scripting before
verification, weakens TLS, or accepts an expired root without a deliberate
offline recovery procedure. Non-interactive apply requires `--yes`; otherwise
the command prints the plan and exits. Interrupted updates leave either the old
verified version or the new verified version active.

Trusted-time policy, expiry tolerance, root thresholds, key custody, delegation,
rotation, revocation, transparency identity, mirror origins, and compromise
recovery are frozen in the updater threat model before code ships. Clock
rollback, missing trusted time, expired metadata, or an unrecognized root fails
closed with an inspectable manual recovery path. Platform notarization or code
signing is an additional operating-system trust signal, not an interchangeable
update signature.

### Laboratory commands

The current v0.7 probe is:

```console
fartapp scenario validate testdata/scenarios/atemporal-probe.json
fartapp scenario validate - --format json
fartapp --help
fartapp help law inspect
fartapp help scenario validate
```

It accepts only the provisional strict JSON envelope documented in
[SCENARIO_PROBE.md](SCENARIO_PROBE.md). The broader command family below remains
planned. Root and nested help enumerate every and only current command path.
They identify the permanent v0.6 string oracle separately from experimental
v0.7 probes and label current English text as presentation rather than a shared
language assumption.

```console
fart scenario init -o reference-enclosure.toml
fart scenario validate reference-enclosure.toml
fart law list
fart law inspect earth.continuum.si
fart scenario capabilities reference-enclosure.toml --format json
fart event run reference-enclosure.toml --output run.fart
fart event inspect run.fart --at "1.2 s"
fart event explain run.fart --why regime.choked
fart event provenance run.fart --to consumers/audio
fart event branch run.fart --set exterior.pressure="0 Pa" --output vacuum.fart
fart event sweep reference-enclosure.toml --vary emitter.pressure="105 kPa..800 kPa" --steps 64
fart event compare small.fart large.fart --nondimensional
fart event translate source.fart --target-world hush-3.toml --mode strict
fart event verify run.fart --refine timestep
fart trace replay run.fart --mode evidence
fart event reconstruct run.fart --mode numerical
fart audio render run.fart --lane physical -o emission.wav
fart symphony render run.fart --mode split -o score.wav
fart radio play drift-93-7 --seed 42
fart chill --station drift-93-7 --event-density sparse
fart mcp serve --transport stdio
fart lab run.fart
fart schema scenario
fart completions powershell
```

Command names stay provisional until the typed command and schema RFC is
accepted. Capabilities and output discipline are not provisional.

### Global output contract

Planned global controls include:

```text
--format text|json|jsonl|csv
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
- Numeric parsing is locale-independent and units are explicit where the law
  profile defines them.
- `-` represents stdin or an explicitly requested artifact stream, never a mix
  of binary output and diagnostics.
- Help includes real examples, units, validity limits, exit codes, and recovery
  suggestions.
- Law-inadmissible states or relations fail with exact field paths and are not silently
  clamped into a different scenario.
- Convenience controls expose the explicit scenario diff they produce.
- Every displayed scientific value can reveal its source, applicable unit,
  uncertainty, validity domain, derivation, and provenance path on demand.

Stable exit-code families should distinguish success, command or input error,
unsupported law or regime, numerical failure, verification failure, archive or
filesystem failure, replay mismatch, and cancellation.

### Identity contract

The interface never collapses these identities:

| Identity | Changes when | Stability promise |
| --- | --- | --- |
| Scenario | Normalized laws, scope, measurement interactions, inputs, or seed change | Unit spelling and field order do not change it |
| Record | A new Lab capture or computation is committed | It does not assert source-law time or recurrence |
| Context occurrence identity claims | An identity actually defined by a scoped context or an explicit inter-law composite relation changes | They may be absent; session order cannot create an identity |
| Occurrence result | Laws, measurement interaction, implementation, numerics, or authoritative claims change | Views and presentation cannot change it |
| View | Knowledge, privacy, accessibility, or selection projection changes | It cannot back-react or change the occurrence account |
| Narrative | Resolved world, storylets, or narrative streams change | Camera and terminal width cannot change it |
| Presentation | Language, layout, device, camera, or accessibility changes | No physical claim follows from it |
| Play session | Rules, initial identity, participants, canonical action journal, or branch changes | Transport and subscriber order cannot change it |
| Archive bytes | Serialization or container bytes change | Byte equality is never confused with semantic equality |

`fart provenance` traverses the typed occurrence provenance graph. `fart
explain` presents a declared derivation or causal path, plus a runnable
counterfactual when the law contexts define one. `fart branch` creates a new
scenario with explicit ancestry and never mutates a certified source.

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
- Core local play, science, replay, proof, and media commands need no display
  server, browser, account, or network access. Network operations such as
  `fart update` and explicitly started remote protocol adapters are separate,
  opt-in command families.

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
consumers/score.json
session/journal.jsonl
session/checkpoints.json
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

Serialization order and session-journal order do not imply source-law temporal
order. Records carry the typed ordering, dependency, concurrency, cyclic, or
inapplicable relations declared by their law contexts. A linear history file is
a storage sequence, not a claim that the source occurrence has linear time.

Initial third-party content packs are versioned, declarative data only. They
cannot load native code, access the network, name arbitrary local files, or
bypass law-capability and resource checks.

A `.fartshow` episode bundle requires law-context set, occurrence scope,
realization provenance, record identity, and certificate. It may add a resolved
world, source, culture, narrative streams, presentation beat ordering,
transcript, audio, and other assets only when supported. Presentation ordering
does not assert source-law time. Seeds alone are not sufficient for archival
replay because content packs evolve.

A session save adds the ruleset, initial identity, roles, privacy-safe actor
identifiers, ordered action journal, checkpoint hashes, branch lineage, artifact
identities, and completion receipt. Opaque live play handles and credentials are
never archived.

## Audio, Symphony, and radio

For compatible Earth-acoustic profiles, the normative offline export begins
with deterministic 48 kHz PCM16 RIFF/WAVE. The pure audio core emits sample
buffers from applicable source terms, a documented stochastic closure, explicit
stream keys, exact rational sample clock, enclosure model, limiter,
quantization, and dither policy. A generic occurrence has no audio requirement.

The TUI may offer optional audio-device playback. Offline WAV must work without
an audio device. Godot consumes the same synthesized frames in real time. Device
resampling is presentation behavior and does not redefine the certified WAV.

Diagnostic sonification declares its units, calibration, mapping, clipping,
quantization, and information loss. Symphony Mode produces a semantic score
from event features. Radio is an independent station-pack presentation layer.
Their identities, controls, manifests, and scientific boundaries are defined in
[AUDIO.md](AUDIO.md).

## Interface 2: Terminal Lab

The TUI should feel like htop for an absurd research instrument. It ships in the
single `fart` binary as `fart lab`, while remaining a separate crate over the
same services.

Generic views include only supported occurrence, participant, coupling,
measurement, view, comparison, invariant, uncertainty, provenance, solver, and
proof concepts. A timeline appears only when a linear ordering exists;
otherwise the TUI uses a relation or dependency view. A capability-driven
registry adds Earth discharge panes for emitter,
interface, plume, payload, acoustics, and conservation, plus compatible
Symphony, radio, agent, spectator, translation, and Broadcast views. An
inapplicable capability removes the pane rather than showing an empty Earth
instrument or fake zero.

Responsive breakpoints:

- `160x48`: wide plots plus applicable ledger, relation, dependency, timeline,
  and narrative views.
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
- Every canonical intent uses the shared service contract and can display or
  copy a CLI representation when one is meaningful. Presentation-only focus and
  layout gestures stay presentation-only.

## Interface 3: Native Lab

The native Godot application ships as a normal Windows, macOS, and Linux app.
It adds capability-selected real-time projections, spatial and procedural audio,
haptics, environments, particles, deposition, and consequence systems while
consuming the same occurrence account and service layer. A 3D view is labeled
as a projection whenever it is not the occurrence's native structure.

Native quality includes keyboard, mouse, controller, remapping, accessibility
setup before first play, safe saves, crash handling, clean uninstall, native
packaging, signing, notarization where applicable, and deterministic archive
export that the CLI can reproduce.

An opt-in, visibly indicated automation mode lets visual agents use captured
frames, synchronized platform accessibility semantics, focus, keyboard,
pointer, controller, and accessibility-invoke operations. It exercises the
rendered application and declared user-facing input path, including applicable
assistive input, rather than calling gameplay internals. It uses inherited
standard I/O where practical and never opens an
always-listening local port.

The first native environment is the biology-neutral Reference Enclosure. The
Tiled Chamber is an optional authored Earth realization. Before native
implementation begins, the same scenario must already run, verify, replay,
translate, render audio, and present in CLI and TUI. The native milestone proves
that every audiovisual and haptic consumer cites the same archive-compatible
occurrence provenance.

## Cross-platform release gate

Every merge eventually enforces:

- Formatting, static analysis, tests, and race or sanitizer checks appropriate to
  each language. Aggregate non-generated core statement coverage remains at
  least 90 percent and every non-generated package remains above 80 percent.
- Domain, solver, archive, protocol, C ABI, and changed high-risk code use
  stricter branch, property, fuzz, mutation, differential, and native-kernel
  evidence. C++ device code, Godot glue, and shaders publish the coverage or
  behavioral surrogate their toolchain can enforce rather than disappearing
  from the quality report.
- Windows, macOS, and Linux replay tolerances.
- Dependency advisory, license, source, and duplicate-version policy.
- Schema validation, archive fuzzing, cancellation and atomic-write fault tests,
  secret scanning, and generated-artifact cleanliness.
- TUI snapshots and terminal restoration smoke tests.
- Headless Godot extension smoke tests once native work begins.
- Direct-service, CLI JSONL, MCP, A2A, TUI, and native same-trace conformance.
- Knowledge-policy, role, idempotency, retry, cancellation, and observation-leak
  tests.
- Official MCP and A2A conformance reports for every advertised protocol and
  binding.

Release archives include executables, licenses, README, completions, manual
pages, schemas, checksums, SBOM, and provenance. macOS uses Developer ID,
hardened runtime, notarization, and stapling. Windows uses Authenticode with a
trusted timestamp. Linux support names its glibc baseline and package format
instead of claiming one build runs everywhere.

The native-start gate requires stable public CLI and TUI releases, a certified
Reference Enclosure archive, optional Tiled Chamber pack, procedural WAV export,
cross-platform QA, and a GDExtension parity test.

## Product identity

The CLI makes the science and comedy complete. The Terminal Lab makes the event
legible in motion. The native app makes it visceral. None is a disposable
wrapper around another.

The research sources are listed in [RESEARCH.md](RESEARCH.md).
