# Security policy

## Supported version

Security fixes are applied to the current default branch. The repository is in
early development, so no older release line is currently supported.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting for this repository. Do not open a
public issue for a suspected vulnerability, exposed secret, malicious archive,
privacy leak, or unsafe content-generation path.

Include:

- The affected commit, file, command, archive, or interface.
- Reproduction steps or a clear source-to-sink trace.
- Expected and observed behavior.
- Impact and required attacker capabilities.
- Any suggested remediation or regression test.

Do not include real credentials, personal data, or harmful payloads when a safe
redacted reproducer is sufficient.

## Security boundaries

The current implementation is a local Go CLI with no network, archive, plugin,
or native application surface. The simulation archive, generated episodes, Rust
core, terminal UI, and Godot application are planned systems. Their documented
security requirements are design gates until corresponding code exists.

Future archive and content-pack readers treat all imported material as untrusted.
Future releases must preserve offline operation, explicit filesystem writes,
bounded resource use, deterministic provenance, and no embedded browser or local
web-service architecture.
