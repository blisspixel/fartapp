# Agent integration status and protocol decisions

An agent can use the current scientific CLI through the portable
[Agent Plugins 1.0.0](https://agent-plugins.org/specification) package rooted at
[plugin.json](../plugin.json). The package supplies the
[fartapp-lab skill](../skills/fartapp-lab/SKILL.md) and a bounded
[recipe corpus](../skills/fartapp-lab/recipes.json). It requires a local command
runner and either Go or a matching executable. Installing the package does not
install a compiler, grant execution permission, or start a service.

This is useful today for law inspection, scenario validation, and the explicit
reservoir, restriction, and coupled-blowdown oracles. Each numerical report
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
adapters; none is a runtime dependency of this Go oracle.

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

## Why service adapters follow the oracle

The planned [agent-play contract](AGENT_PLAY.md) requires stable canonical
actions and observations, roles, budgets, journals, task controls, and retained
artifacts. These are still design obligations. Wrapping the present CLI in a
server would not establish those semantics. The current package makes the
implemented commands usable while preserving that boundary.

Both future adapters call the same in-process service as the CLI. They do not
shell out to this executable or infer authoritative state from its text. The
first adapter milestone must earn all of the following:

- Direct-service and machine-CLI parity for each advertised operation, including
  authorization, refusal, idempotency, budgets, and artifact provenance.
- MCP revision discovery and negotiation, clean bounded stdio framing,
  result metadata, and HTTP routing and authorization behavior for the selected
  revision. Legacy compatibility must be tested separately if offered.
- Explicitly negotiated [MCP Tasks](https://modelcontextprotocol.io/extensions/tasks/overview)
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
An SDK version or a portable plugin manifest cannot promote the unfinished
play service to a supported capability.
