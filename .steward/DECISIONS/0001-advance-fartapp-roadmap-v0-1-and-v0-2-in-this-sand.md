# 0001. Advance fartapp roadmap v0.1 and v0.2 in this sandbox worktree only.

v0.1 intensity parsing: main.go must require one CLI argument, parse it as an integer strictly from 1 to 5 inclusive, and reject anything else (missing arg, non-integer, out of range) with a clear stderr message and exit code 1.

v0.2 emission mapping: fix Pick in fart.go so intensity 1 maps to sounds[0] through intensity 5 to sounds[4] (sounds[intensity-1] after validation).

Add tests in fart_test.go: TestPick for all five intensities, plus tests covering invalid CLI input behavior.

Keep the change minimal and gofmt-clean. Match README example: fartapp 3 prints braaap (respectable). No AI attribution in any file.

- **Status:** executed (tests pass)
- **Decided:** 2026-06-21T22:47:25Z
- **Verification:** proof gate: pass; differential: sustained (baseline pass → after pass); tests 1 → 1; confidence 0.45 - basis: +0.50 re-executed verification passed (harness ran it, not the worker's word); +0.10 baseline also passed: no regression introduced (not proof of new behavior); -0.20 independent checker's last verdict was a rejection; +0.05 acceptance criteria: all 4 re-executed criterion checks met
- **Evidence:** 7ba33fbea03d1d4c4a70c166fd16f03c1aee58ca009bbca05fca2d8603578e40

## Why

Approved and executed through the proof-gated loop. Goal: Advance fartapp roadmap v0.1 and v0.2 in this sandbox worktree only.

v0.1 intensity parsing: main.go must require one CLI argument, parse it as an integer strictly from 1 to 5 inclusive, and reject anything else (missing arg, non-integer, out of range) with a clear stderr message and exit code 1.

v0.2 emission mapping: fix Pick in fart.go so intensity 1 maps to sounds[0] through intensity 5 to sounds[4] (sounds[intensity-1] after validation).

Add tests in fart_test.go: TestPick for all five intensities, plus tests covering invalid CLI input behavior.

Keep the change minimal and gofmt-clean. Match README example: fartapp 3 prints braaap (respectable). No AI attribution in any file.
