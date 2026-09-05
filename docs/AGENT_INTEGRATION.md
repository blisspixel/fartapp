# Agent integration status and protocol decisions

An agent can use the current scientific CLI through the portable
[Agent Plugins 1.0.0](https://agent-plugins.org/specification) package rooted at
[plugin.json](../plugin.json). The package supplies the
[fartapp-lab skill](../skills/fartapp-lab/SKILL.md) and a bounded
[recipe corpus](../skills/fartapp-lab/recipes.json). It requires a local command
runner and either Go or a matching executable. Installing the package does not
install a compiler, grant execution permission, or start a service.

This is useful today for assurance and law inspection, scenario validation, and
the explicit reservoir, restriction, and coupled-blowdown oracles. Assurance
inspection exposes declared metadata without executing checks. Each numerical report
retains its assumptions and evidence limits. The skill directs agents to inspect
the exit code as well as JSON, preserve refused inputs, and distinguish
arithmetic certification from scientific validation.

## Package and executable evidence

The repository itself is the plugin root. It contains the portable root
manifest and an immediate `skills/fartapp-lab/SKILL.md` child. Fixtures, recipes,
and referenced documentation ship inside that root. The manifest targets the
published 1.0.0 schema; the 1.1 draft is not an implicit dependency.

The package has no `mcp.json` because it does not supply an MCP server. Its
version is an experimental package release, not a promise of stable scientific
wire formats. The recipe corpus covers successful inspection and calculation
as well as refusal of an incompatible law context. Executable tests check
stdout, stderr, exit status, selected JSON fields, and file versus stdin parity.
The repository policy checker verifies this package's bounded manifest,
skill, and recipe profile offline, including containment of package paths.
It is not a general Agent Plugins client-conformance implementation.

The package includes an executable refinement recipe with separate tolerance
and completion assertions, plus a metadata-only assurance recipe. Its retention
instructions point to the
[evidence carrier](WALK_EVIDENCE.md), whose CLI tests exercise actual capture,
replay, corruption refusal, and reconstruction mismatch. All 13 recipe argument
arrays execute the Go CLI. The Rust CLI has separate complete-report parity
gates for reservoir, restriction, and prescribed-area history services
and an [eight-command play trace](../testdata/play/reservoir-session.jsonl)
checked through both the direct service and native CLI.

The native [play session](PLAY_SESSION.md) now supplies a bounded local profile
for explicit reservoir experiments. Its JSONL transport, retained transcript,
integrity replay, and explicit fresh reconstruction require no agent host.
Reconstruction reports exact current-implementation agreement independently
of retained session completion, with all fresh reports available on mismatch.
The skill distinguishes this Rust
command surface from its Go recipe corpus. A local session reference is a
reproducible fingerprint, not an authenticated or globally fresh authority token.

The [walkthrough](WALKTHROUGH.md) adds an end-to-end investigation: simulate a
low-pressure case, compare outlet area, retain a witness, and reconstruct
against that retained expectation. The witness includes the implementation and
runtime profile. It is not a cross-compiler numerical parity guarantee,
canonical case identity, signed archive, or certificate authority.

A separate documentation-guided agent exercise on Go 1.27.1, Windows/amd64
completed that workflow. It obtained the same mass endpoint and half the
numerical elapsed time after doubling the outlet area, reconstructed a retained
witness successfully, and rejected that witness for the changed-area case
while preserving both digests. This checks the skill's usability with a local
command runner; it is not an agent-host installation test.

Host installation and command permissions remain client-specific. A portable
manifest does not establish that every client's native plugin format is the
same. This release's evidence is package validation and actual CLI execution;
installation in individual agent hosts has not been verified. Consult the
[compatible-client directory](https://agent-plugins.org/compatible-clients)
and the selected client's installation documentation before making a host
compatibility claim.

## Reviewed protocol baseline

The following versions were reviewed on 2026-09-04. They guide the future
adapters; none is a runtime dependency of the Go or native Rust application.

| Component | Reviewed baseline | Decision |
| --- | --- | --- |
| Agent Plugins | Published 1.0.0 | Ship the portable skill package now |
| MCP | [2026-07-28](https://modelcontextprotocol.io/specification/2026-07-28) | Target the current revision explicitly; do not copy the old initialization and session model into new code |
| MCP Go SDK | [v1.7.0](https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.7.0) | Available for a Go adapter if its service boundary is retained; review stateless HTTP configuration and negotiated revision |
| MCP Rust SDK | [rmcp 3.2.0](https://github.com/modelcontextprotocol/rust-sdk/releases/tag/rmcp-v3.2.0) | Stable SDK candidate for the planned Rust service, subject to conformance tests |
| A2A specification | [v1.0.1](https://github.com/a2aproject/A2A/releases/tag/v1.0.1), wire `1.0` | Separate agent coordination from tool access |
| A2A Go SDK | [v2.5.0](https://github.com/a2aproject/a2a-go/releases/tag/v2.5.0) | Research baseline, not an added dependency |

Agent Plugins packages skills and MCP connection configuration. It does not
define an A2A transport or replace either protocol. A2A task identity, an MCP
request, and a Lab play handle serve different purposes.

## External harness research

The following primary sources were reviewed on 2026-09-05. These are external
orchestrator candidates and useful design references. No installation or
end-to-end compatibility test with these hosts has been performed. Their release
labels do not establish this package's compatibility, and none is an application
runtime dependency.

| Project | Reviewed identity and maturity | Useful boundary |
| --- | --- | --- |
| Hermes Agent | Nous Research's agent harness, [v0.21.0 / tag v2026.8.31](https://github.com/NousResearch/hermes-agent/releases/tag/v2026.8.31) | Its [MCP client and separate session server](https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp) and optional [ACP integration](https://hermes-agent.nousresearch.com/docs/user-guide/features/acp) are distinct interfaces |
| Pi | The coding agent now maintained at `earendil-works/pi`, [v0.85.0](https://github.com/earendil-works/pi/releases/tag/v0.85.0) | Its [pinned command and extension contract](https://raw.githubusercontent.com/earendil-works/pi/v0.85.0/packages/coding-agent/README.md) exposes RPC and SDK modes but no built-in MCP support; Pi RPC is not MCP, ACP, or A2A |
| OpenClaw | The OpenClaw agent/Gateway runtime, stable release [v2026.9.1](https://github.com/openclaw/openclaw/releases/tag/v2026.9.1) | Its [security model](https://docs.openclaw.ai/gateway/security) assumes one trusted operator boundary per Gateway; session routing keys are not authentication |
| NemoClaw | NVIDIA's OpenShell-based reference stack, reviewed source [v0.0.120](https://github.com/NVIDIA/NemoClaw/tree/v0.0.120) | The official [platform policy](https://docs.nvidia.com/nemoclaw/latest/user-guide/hermes/reference/platform-support) identifies an alpha early preview; it is not a GA or universal platform-support claim |
| DeepSeek Harness | DeepSeek AI's open-source harness, currently [developer preview](https://github.com/deepseek-ai/deepseek-harness) with announced compatibility-breaking changes | Its [plugin composition and append-only session views](https://www.deepseek.com/harness/en/) are useful design references; its plugin system is not the Agent Plugins package format |

Hermes and Pi publish installable releases. This review does not assert a formal
GA guarantee for their integration APIs. OpenClaw's stable release channel is
also separate from any compatibility or multi-tenant isolation guarantee.

Three patterns are useful for this application:

1. Keep the harness outside the scientific application. Reuse the existing
   skill and literal CLI recipes through the host's command runner. Hermes
   supports [external skill directories](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills/),
   Pi supports [skill discovery](https://pi.dev/docs/latest/skills), and OpenClaw
   documents [workspace skills](https://docs.openclaw.ai/tools/skills).
   Discovery is not permission or write protection. The next useful host test
   is the same low-pressure prediction, changed-area comparison, witness
   retention, and reconstruction refusal already tested locally, followed by
   the native play trace. Keep Linux and supported macOS hosts first; Hermes's
   [platform policy](https://hermes-agent.nousresearch.com/docs/getting-started/platform-support)
   distinguishes Apple Silicon support from unsupported Intel macOS.
2. Retain domain evidence separately from conversation projections. DeepSeek's
   session-log views and Pi's parent-linked sessions suggest useful inspection
   patterns, but conversation summaries do not become numerical evidence.
   Our accepted receipts bind normalized requests, reports, and journal order.
   Pi's [RPC event contract](https://pi.dev/docs/latest/rpc) also distinguishes
   an agent turn ending from a fully settled run. Transfer that distinction to
   explicit completion and cancellation semantics rather than reusing host
   events as Lab state. The native transport already uses bounded LF framing;
   Unicode line separators remain JSON data.
3. Keep authentication, credentials, and sandbox policy outside scientific
   identity. OpenClaw's trust boundary reinforces the difference between a
   routing reference and authority. NemoClaw's
   [managed MCP policy](https://docs.nvidia.com/nemoclaw/user-guide/openclaw/manage-sandboxes/mcp-servers/about-managed-mcp-servers)
   keeps provider credentials outside the sandbox and constrains endpoints,
   paths, adapters, and methods. That managed path supports authenticated HTTPS
   Streamable HTTP, not launching a stdio server. It does not justify adding an
   HTTP listener to this offline application. Host prompts, skills, and retained
   reports must not grant themselves additional tool access.

These patterns require no embedded agent loop, model provider, JavaScript or
Python runtime, automatic host installation, or new network service. A host's
offline startup option also does not constrain every provider or tool request;
offline operation must be checked at the actual process and network boundary.

## Why service adapters follow the oracle

The [local reservoir profile](PLAY_SESSION.md) now implements explicit actions,
read-only observations, an operator binding, attempt accounting, retained
idempotency, and an integrity journal. The broader
[agent-play contract](AGENT_PLAY.md) still requires general case identity,
multiple roles and knowledge policies, asynchronous task controls, and durable
live authority. Wrapping the current CLI in a server would not establish those
semantics.

Both future adapters call the same in-process service as the CLI. They do not
shell out to this executable or infer authoritative state from its text. The
advertised adapter capabilities need the applicable evidence below:

- Direct-service and machine-CLI parity for each advertised operation, including
  authorization, refusal, idempotency, budgets, and artifact provenance.
- MCP revision discovery and negotiation, clean bounded stdio framing,
  result metadata, and HTTP routing and authorization behavior for the selected
  revision. Legacy compatibility must be tested separately if offered.
- If tasks are advertised, explicitly negotiated [MCP Tasks](https://modelcontextprotocol.io/extensions/tasks/overview)
  with tested progress, retention, cancellation, and expiry. Cancelling a
  transport request, cancelling a task, and stopping a play session are
  separate operations.
- An accurate A2A Agent Card, advertised wire version and bindings, bounded
  task/context identifiers, and application-level retry and idempotency rules.
  Production HTTP requires TLS; explicit loopback development is a separate
  mode. No A2A stdio binding is invented.
- Pinned official conformance suites and interoperability cases with executed,
  skipped, failed, and unsupported counts retained. A successful harness exit
  is not evidence that every protocol case ran.

Protocol upgrades pass these tests before their advertised revision changes.
An SDK version or a portable plugin manifest cannot expand the delivered local
profile into unimplemented case, authority, or protocol capabilities.
