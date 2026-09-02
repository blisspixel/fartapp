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

Optional MCP and A2A adapters are interoperability surfaces, not the desktop
application architecture. Local protocol services start only through an
explicit command. MCP defaults to inherited standard I/O. A2A initially binds
to loopback with a random, role-bound capability kept out of arguments, logs,
archives, and shell history.

Non-loopback service requires TLS, authentication, per-session and per-role
authorization, expiry, quotas, concurrency and size limits, cancellation, and a
separate threat review. Agent Cards, invites, task lists, artifacts, and
spectator streams are scoped to the caller. Imported agent text, lore, lyrics,
archives, resource identifiers, and terminal content are untrusted data.

No game or protocol action grants shell access, loads native code, follows an
arbitrary URL, writes an arbitrary path, or reads credentials. Push
notifications remain disabled until webhook authentication, duplicate delivery,
HTTPS validation, SSRF defenses, rate limits, and retry budgets are verified.
