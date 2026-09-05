use std::ffi::OsString;

pub(super) const ROOT: &str = "\
F.A.R.T. Lab experimental native Rust CLI

Usage:
  fart <command> [options]
  fart <intensity>
  fart help [command]

Commands:
  reservoir    Predict an explicit analytical SI endpoint.
  restriction  Predict flow or integrate a prescribed area history.
  play         Run, replay, or reconstruct a bounded local experiment.
  help         Show command help without reading inputs.

Start here:
  fart 3
  fart help reservoir predict
  fart help play run

The permanent toy accepts an integer from 1 to 5.
This experimental build provides no protocol server or empirical validation.
";

pub(super) const RESERVOIR: &str = "\
RESERVOIR PREDICT

Usage:
  fart reservoir predict <request.json|-> [--format text|json]

Predict one rigid ideal-mixture reservoir endpoint from explicit SI inputs.
Use - for stdin. Text is the default; JSON retains full numeric precision.

Example:
  fart reservoir predict testdata/reservoir/synthetic-mixture-adiabatic.json

Limits:
  65,536 input bytes. Duplicate, unknown, and null JSON members are refused.
  Model, units, components, closure, and withdrawal fraction are explicit.
  No elapsed time, case commitment, certificate, or empirical validation.

Exit status:
  0  Prediction completed.
  1  Usage, input, model, or output failure.
     A refused JSON request still emits its structured report.
";

pub(super) const RESTRICTION: &str = "\
RESTRICTION

Usage:
  fart restriction predict <request.json|-> [--format text|json]
  fart restriction history <request.json|-> [--format text|json]

Commands:
  predict  Evaluate a quasi-steady isentropic converging restriction.
  history  Integrate prescribed area samples at a frozen stagnation state.

Use - for stdin. Inputs use explicit SI quantities and model references.
Both operations describe a narrow ideal-gas model without empirical validation.
";

pub(super) const RESTRICTION_PREDICT: &str = "\
RESTRICTION PREDICT

Usage:
  fart restriction predict <request.json|-> [--format text|json]

Predict subsonic, choked, or closed flow from explicit SI stagnation, back
pressure, discharge coefficient, and prescribed or compliant area inputs.
Use - for stdin. Text is the default; JSON retains full numeric precision.

Example:
  fart restriction predict testdata/restriction/gamma15-choked.json

Limits:
  65,536 input bytes. Duplicate, unknown, and null JSON members are refused.
  No reservoir depletion, elapsed time, empirical calibration, or certificate.

Exit status:
  0  Prediction completed, including supported no-flow states.
  1  Usage, input, model, or output failure; JSON retains structured refusals.
";

pub(super) const RESTRICTION_HISTORY: &str = "\
RESTRICTION HISTORY

Usage:
  fart restriction history <request.json|-> [--format text|json]

Integrate 1 to 256 prescribed area samples with strictly increasing times.
Stagnation, back pressure, and discharge coefficient stay explicitly fixed.
Use - for stdin. Text is the default; JSON retains all samples and claims.

Example:
  fart restriction history testdata/restriction/gamma15-choked-history.json

Limits:
  65,536 input bytes. Duplicate, unknown, and null JSON members are refused.
  Totals use trapezoidal integration of the supplied sample values.
  A single sample has zero elapsed duration and zero integrated totals.
  No coupled blowdown, hidden time steps, empirical validation, or certificate.

Exit status:
  0  Prescribed history completed.
  1  Usage, input, model, or output failure; JSON retains structured refusals.
";

pub(super) const PLAY: &str = "\
LOCAL RESERVOIR PLAY

Usage:
  fart play run <commands.jsonl|-> [--format text|jsonl|transcript]
  fart play replay <transcript.json|-> [--format text|json]
  fart play reconstruct <transcript.json|-> [--format text|json]

Commands:
  run     Start, predict, observe, and explicitly finish one local session.
  replay  Verify and project retained evidence without running the solver.
  reconstruct  Recompute every retained attempt and compare exact values.

Help:
  fart help play run
  fart help play replay
  fart help play reconstruct
";

pub(super) const PLAY_RUN: &str = "\
PLAY RUN

Usage:
  fart play run <commands.jsonl|-> [--format text|jsonl|transcript]

Read one strict JSON command per LF-terminated line; CRLF is also accepted.
Use - for stdin. The final line may end at EOF. Blank lines are invalid commands.

Formats:
  text        Readable replies and an explicit end-of-input summary (default).
  jsonl       One reply per command, then one summary with retained transcript.
  transcript  Only the exact retained transcript at EOF, for play replay.

Example:
  fart play run testdata/play/reservoir-session.jsonl
  fart play run testdata/play/reservoir-session.jsonl --format transcript

Limits:
  64 KiB per command, 1 MiB total input, 128 commands, 16 MiB total output.
  Start declares at most 16 prediction attempts. Reads and exact retries
  cost no attempts. Every prediction uses the original authored baseline.

Exit status:
  0  Input delivered with no rejected commands and an explicit finish.
  1  Rejected command, incomplete session, usage, transport, or output failure.
     A costed model refusal is retained evidence, not a transport rejection.
     EOF never invents a finish. An unfinished transcript can still be retained.
";

pub(super) const PLAY_REPLAY: &str = "\
PLAY REPLAY

Usage:
  fart play replay <transcript.json|-> [--format text|json]

Verify the retained structure and digest chain, then project its account.
This does not rerun predictions, restore live authority, or prove authorship.
Use - for stdin. Input is limited to 8 MiB; text is the default.

Example:
  fart play replay artifacts/session.json

Create that input by redirecting play run --format transcript to a file.
A replay accepts a transcript object, not the mixed reply/summary JSONL stream.

Exit status:
  0  Retained integrity verified, including honestly unfinished sessions.
  1  Usage, input, integrity, or output failure.
";

pub(super) const PLAY_RECONSTRUCT: &str = "\
PLAY RECONSTRUCT

Usage:
  fart play reconstruct <transcript.json|-> [--format text|json]

Verify the entire retained chain, freshly admit its baseline, and recompute
every costed prediction attempt, including model refusals. The exact canonical
comparison covers all reports and receipts using the current implementation.
Use - for stdin. Input is limited to 8 MiB; output to 16 MiB. Text is the default.

Example:
  fart play reconstruct artifacts/session.json --format json

Create that input by redirecting play run --format transcript to a file.
JSON retains the original summary and all fresh evidence for investigation.
The original is never modified; this does not restore a live session.
Agreement is not empirical validation, authentication, or a promise of
identical floating-point results across platforms or implementations.

Exit status:
  0  Exact current-implementation match, including unfinished sessions.
  1  Mismatch, refused admission, usage, input, integrity, or output failure.
";

pub(super) fn topic(args: &[OsString]) -> Option<&'static str> {
    let words: Vec<_> = args.iter().map(|arg| arg.to_str()).collect();
    match words.as_slice() {
        [] => Some(ROOT),
        [Some("reservoir")] | [Some("reservoir"), Some("predict")] => Some(RESERVOIR),
        [Some("restriction")] => Some(RESTRICTION),
        [Some("restriction"), Some("predict")] => Some(RESTRICTION_PREDICT),
        [Some("restriction"), Some("history")] => Some(RESTRICTION_HISTORY),
        [Some("play")] => Some(PLAY),
        [Some("play"), Some("run")] => Some(PLAY_RUN),
        [Some("play"), Some("replay")] => Some(PLAY_REPLAY),
        [Some("play"), Some("reconstruct")] => Some(PLAY_RECONSTRUCT),
        _ => None,
    }
}
