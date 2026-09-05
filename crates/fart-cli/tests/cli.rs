//! Native process behavior, bounded I/O, and presentation failure contracts.

use std::ffi::OsString;
use std::fs;
use std::io::{self, Read, Write};
use std::path::PathBuf;
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicU64, Ordering};

use fart_services::MAX_INPUT_BYTES;
use serde_json::Value;

const FIXTURE: &[u8] =
    include_bytes!("../../../testdata/reservoir/synthetic-mixture-adiabatic.json");
const PLAY_SCRIPT: &[u8] = include_bytes!("../../../testdata/play/reservoir-session.jsonl");

fn run(args: &[&str], data: &[u8]) -> (u8, Vec<u8>, Vec<u8>) {
    let args: Vec<_> = args.iter().map(OsString::from).collect();
    let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
    let code = fart_cli::run(&args, &mut &data[..], &mut stdout, &mut stderr);
    (code, stdout, stderr)
}

struct TempDirectory(PathBuf);
impl TempDirectory {
    fn new() -> Self {
        static NEXT: AtomicU64 = AtomicU64::new(0);
        let path = std::env::temp_dir().join(format!(
            "fart-rust-cli-{}-{}",
            std::process::id(),
            NEXT.fetch_add(1, Ordering::Relaxed)
        ));
        fs::create_dir(&path).unwrap();
        Self(path)
    }
}
impl Drop for TempDirectory {
    fn drop(&mut self) {
        let _ = fs::remove_dir_all(&self.0);
    }
}

#[test]
fn permanent_toy_native_stdout_and_exit_status_are_exact() {
    for (intensity, expected) in [
        ("1", "pfft (gentle)\n"),
        ("2", "toot (respectable)\n"),
        ("3", "braaap (respectable)\n"),
        ("4", "blorp (respectable)\n"),
        ("5", "KABLAM (mighty)\n"),
    ] {
        let result = Command::new(env!("CARGO_BIN_EXE_fart"))
            .arg(intensity)
            .output()
            .unwrap();
        assert!(result.status.success());
        assert_eq!(result.stdout, expected.as_bytes());
        assert!(result.stderr.is_empty());
    }
    for invalid in ["0", "6", "01", "+1", "-1", "1.0", " 1", "\u{ff11}"] {
        let (code, stdout, stderr) = run(&[invalid], b"");
        assert_eq!(code, 1);
        assert!(stdout.is_empty());
        assert!(
            String::from_utf8(stderr)
                .unwrap()
                .contains("canonical integer")
        );
    }
}

#[test]
fn all_help_routes_are_explicit_and_do_not_read_input() {
    for args in [
        vec!["--help"],
        vec!["-h"],
        vec!["help"],
        vec!["help", "reservoir"],
        vec!["help", "reservoir", "predict"],
        vec!["reservoir", "--help"],
        vec!["reservoir", "predict", "--help"],
        vec!["help", "play"],
        vec!["help", "play", "run"],
        vec!["help", "play", "replay"],
        vec!["play", "--help"],
        vec!["play", "run", "--help"],
        vec!["play", "replay", "--help"],
        vec!["play", "reconstruct", "--help"],
        vec!["help", "play", "reconstruct"],
        vec!["restriction", "--help"],
        vec!["restriction", "predict", "--help"],
        vec!["restriction", "history", "--help"],
        vec!["help", "restriction"],
        vec!["help", "restriction", "predict"],
        vec!["help", "restriction", "history"],
    ] {
        let arguments: Vec<_> = args.iter().map(OsString::from).collect();
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        let mut reader = FailingReader;
        assert_eq!(
            fart_cli::run(&arguments, &mut reader, &mut stdout, &mut stderr),
            0
        );
        let text = String::from_utf8(stdout).unwrap();
        assert!(text.contains("Usage:"));
        if args.contains(&"reservoir") {
            assert!(text.starts_with("RESERVOIR PREDICT\n"));
            assert!(text.contains("65,536"));
        } else if args.contains(&"restriction") {
            assert!(text.starts_with("RESTRICTION"));
            assert!(text.contains("restriction"));
        } else if args.contains(&"play") {
            assert!(text.contains("play"));
            assert!(!text.contains("permanent toy"));
        } else {
            assert!(text.contains("experimental native Rust CLI"));
        }
        assert!(stderr.is_empty());
    }
}

#[test]
fn usage_errors_do_not_produce_success_output() {
    for args in [
        vec![],
        vec!["1", "2"],
        vec!["help", "unknown"],
        vec!["reservoir"],
        vec!["reservoir", "simulate"],
        vec!["reservoir", "predict"],
        vec!["reservoir", "predict", "-", "--format"],
        vec!["reservoir", "predict", "-", "--format", "yaml"],
        vec![
            "reservoir",
            "predict",
            "-",
            "--format",
            "json",
            "--format",
            "text",
        ],
        vec!["reservoir", "predict", "-", "--help", "--help"],
        vec!["reservoir", "predict", "-", "--unknown"],
        vec!["reservoir", "predict", "one.json", "two.json"],
        vec!["play"],
        vec!["play", "unknown"],
        vec!["play", "run"],
        vec!["play", "replay"],
        vec!["play", "run", "-", "--format", "json"],
        vec!["play", "replay", "-", "--format", "jsonl"],
        vec!["play", "run", "-", "--format", "text", "--format", "text"],
        vec!["play", "run", "--help", "--help"],
        vec!["play", "run", "-", "--bad"],
        vec!["play", "run", "one", "two"],
        vec!["play", "reconstruct"],
        vec!["play", "reconstruct", "-", "--format", "transcript"],
        vec!["restriction"],
        vec!["restriction", "unknown"],
        vec!["restriction", "predict"],
        vec!["restriction", "history"],
        vec!["restriction", "history", "-", "--format", "jsonl"],
        vec!["restriction", "predict", "one", "two"],
    ] {
        let (code, stdout, stderr) = run(&args, FIXTURE);
        assert_eq!(code, 1, "{args:?}");
        assert!(stdout.is_empty(), "{args:?}");
        assert!(!stderr.is_empty(), "{args:?}");
    }
}

#[test]
fn file_stdin_and_native_process_reports_agree_without_modifying_the_input() {
    let directory = TempDirectory::new();
    let source = directory.0.join("case with spaces.json");
    fs::write(&source, FIXTURE).unwrap();
    let mut args = vec![
        OsString::from("reservoir"),
        OsString::from("predict"),
        source.into_os_string(),
        OsString::from("--format"),
        OsString::from("json"),
    ];
    let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
    assert_eq!(
        fart_cli::run(&args, &mut &b""[..], &mut stdout, &mut stderr),
        0
    );
    assert!(stderr.is_empty());
    assert_eq!(fs::read(&args[2]).unwrap(), FIXTURE);
    let (code, streamed, stderr) = run(&["reservoir", "predict", "-", "--format", "json"], FIXTURE);
    assert_eq!(code, 0);
    assert_eq!(streamed, stdout);
    assert!(stderr.is_empty());
    let value: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(value["status"], "predicted");
    assert_eq!(value["final"]["temperature_k"], 200.0);
    args[2] = OsString::from("-");
    let mut process = Command::new(env!("CARGO_BIN_EXE_fart"))
        .args(&args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    process.stdin.take().unwrap().write_all(FIXTURE).unwrap();
    let result = process.wait_with_output().unwrap();
    assert!(result.status.success());
    assert_eq!(result.stdout, stdout);
    assert!(result.stderr.is_empty());
    let (code, text, stderr) = run(
        &["reservoir", "predict", "--format", "text", "--", "-"],
        FIXTURE,
    );
    assert_eq!(code, 0);
    assert!(
        String::from_utf8(text)
            .unwrap()
            .contains("RESERVOIR ENDPOINT PREDICTED")
    );
    assert!(stderr.is_empty());
}

#[test]
fn refused_json_and_text_keep_their_distinct_stream_contracts() {
    let (code, stdout, stderr) = run(&["reservoir", "predict", "-", "--format", "json"], b"{}");
    assert_eq!(code, 1);
    assert!(stderr.is_empty());
    let value: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(value["status"], "invalid");
    assert!(value.get("final").is_none());
    let (code, stdout, stderr) = run(&["reservoir", "predict", "-"], b"{}");
    assert_eq!(code, 1);
    assert!(stdout.is_empty());
    assert!(
        String::from_utf8(stderr)
            .unwrap()
            .contains("unsupported_schema")
    );
    let directory = TempDirectory::new();
    let missing = directory.0.join("missing.json");
    let result = Command::new(env!("CARGO_BIN_EXE_fart"))
        .args([
            OsString::from("reservoir"),
            OsString::from("predict"),
            missing.into_os_string(),
            OsString::from("--format"),
            OsString::from("json"),
        ])
        .output()
        .unwrap();
    assert_eq!(result.status.code(), Some(1));
    assert!(result.stderr.is_empty());
    let value: Value = serde_json::from_slice(&result.stdout).unwrap();
    assert_eq!(value["diagnostics"][0]["reason_code"], "input_not_found");
}

struct FailingReader;
impl Read for FailingReader {
    fn read(&mut self, _: &mut [u8]) -> io::Result<usize> {
        Err(io::Error::other("unavailable"))
    }
}

struct FailingWriter;
impl Write for FailingWriter {
    fn write(&mut self, _: &[u8]) -> io::Result<usize> {
        Err(io::Error::from(io::ErrorKind::BrokenPipe))
    }
    fn flush(&mut self) -> io::Result<()> {
        Ok(())
    }
}

struct CountedReader {
    read: usize,
}
impl Read for CountedReader {
    fn read(&mut self, bytes: &mut [u8]) -> io::Result<usize> {
        bytes.fill(b' ');
        self.read += bytes.len();
        Ok(bytes.len())
    }
}

#[test]
fn read_budget_and_io_failures_are_observable() {
    let args = ["reservoir", "predict", "-", "--format", "json"].map(OsString::from);
    let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
    let mut reader = CountedReader { read: 0 };
    assert_eq!(
        fart_cli::run(&args, &mut reader, &mut stdout, &mut stderr),
        1
    );
    assert_eq!(reader.read, MAX_INPUT_BYTES + 1);
    let value: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(value["diagnostics"][0]["reason_code"], "input_too_large");
    stdout.clear();
    assert_eq!(
        fart_cli::run(&args, &mut FailingReader, &mut stdout, &mut stderr),
        1
    );
    let value: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(value["diagnostics"][0]["reason_code"], "input_unavailable");
    assert_eq!(
        fart_cli::run(&args, &mut &FIXTURE[..], &mut FailingWriter, &mut stderr),
        1
    );
    assert!(String::from_utf8(stderr).unwrap().contains("write output"));
    assert_eq!(
        fart_cli::run(
            &[OsString::from("3")],
            &mut &b""[..],
            &mut FailingWriter,
            &mut FailingWriter
        ),
        1
    );
}

#[test]
fn play_cli_and_direct_service_return_the_same_literal_trace_and_evidence() {
    use fart_services::play::PlayService;
    let mut service = PlayService::new();
    let mut expected = String::new();
    for command in std::str::from_utf8(PLAY_SCRIPT).unwrap().lines() {
        expected.push_str(&service.process_json(command.as_bytes()).to_json());
        expected.push('\n');
    }
    let summary = service.end_of_input();
    expected.push_str(&summary.to_json());
    expected.push('\n');
    let args = ["play", "run", "-", "--format", "jsonl"];
    let (code, stdout, stderr) = run(&args, PLAY_SCRIPT);
    assert_eq!(code, 0);
    assert_eq!(stdout, expected.as_bytes());
    assert!(stderr.is_empty());
    let rows: Vec<Value> = expected
        .lines()
        .map(|line| serde_json::from_str(line).unwrap())
        .collect();
    assert_eq!(rows.len(), 9);
    assert_eq!(rows[1], rows[3]);
    assert_eq!(rows[1], rows[7]);
    assert_eq!(rows[8]["complete"], true);
    assert_eq!(rows[8]["revision"], 3);
    assert_eq!(rows[8]["attempts_remaining"], 2);

    let directory = TempDirectory::new();
    let path = directory.0.join("play commands.jsonl");
    fs::write(&path, PLAY_SCRIPT).unwrap();
    let result = Command::new(env!("CARGO_BIN_EXE_fart"))
        .args(["play", "run"])
        .arg(&path)
        .args(["--format", "jsonl"])
        .output()
        .unwrap();
    assert!(result.status.success());
    assert_eq!(result.stdout, stdout);
    assert!(result.stderr.is_empty());
    assert_eq!(fs::read(path).unwrap(), PLAY_SCRIPT);
    let (code, transcript, stderr) =
        run(&["play", "run", "-", "--format", "transcript"], PLAY_SCRIPT);
    assert_eq!(code, 0);
    assert!(stderr.is_empty());
    assert_eq!(
        std::str::from_utf8(&transcript).unwrap().trim_end(),
        summary.transcript().unwrap().to_json()
    );
    for format in ["text", "json"] {
        let (code, replayed, stderr) =
            run(&["play", "replay", "-", "--format", format], &transcript);
        assert_eq!(code, 0);
        assert!(!replayed.is_empty());
        assert!(stderr.is_empty());
    }
    let (code, text, stderr) = run(&["play", "run", "-"], PLAY_SCRIPT);
    assert_eq!(code, 0);
    assert!(stderr.is_empty());
    assert!(
        std::str::from_utf8(&text)
            .unwrap()
            .contains("PLAY RUN FINISHED")
    );
}

#[test]
fn play_eof_preserves_unfinished_evidence_and_replay_is_read_only() {
    let unfinished = std::str::from_utf8(PLAY_SCRIPT)
        .unwrap()
        .lines()
        .take(6)
        .collect::<Vec<_>>()
        .join("\n");
    let (code, retained, stderr) = run(
        &["play", "run", "-", "--format", "transcript"],
        unfinished.as_bytes(),
    );
    assert_eq!(code, 1);
    assert!(stderr.is_empty());
    let value: Value = serde_json::from_slice(&retained).unwrap();
    assert_eq!(value["summary"]["complete"], false);
    assert_eq!(value["summary"]["revision"], 2);
    let before = retained.clone();
    let (code, projected, stderr) = run(&["play", "replay", "-", "--format", "json"], &retained);
    assert_eq!(code, 0);
    assert!(stderr.is_empty());
    let projected: Value = serde_json::from_slice(&projected).unwrap();
    assert_eq!(projected["complete"], false);
    assert_eq!(retained, before);
    let (code, stdout, stderr) = run(&["play", "run", "-", "--format", "transcript"], b"");
    assert_eq!(code, 1);
    assert!(stdout.is_empty());
    assert!(
        std::str::from_utf8(&stderr)
            .unwrap()
            .contains("never started")
    );

    for format in ["text", "json"] {
        let (code, stdout, stderr) = run(&["play", "replay", "-", "--format", format], b"{}");
        assert_eq!(code, 1);
        assert_eq!(stdout.is_empty(), format == "text");
        assert_eq!(stderr.is_empty(), format == "json");
    }
    let (code, stdout, stderr) = run(&["play", "replay", "missing-play-evidence.json"], b"");
    assert_eq!(code, 1);
    assert!(stdout.is_empty());
    assert!(
        std::str::from_utf8(&stderr)
            .unwrap()
            .contains("check the path")
    );
}

#[test]
fn native_reconstruction_matches_direct_evidence_and_keeps_unfinished_status_separate() {
    for count in [1, 6, 8] {
        let script = std::str::from_utf8(PLAY_SCRIPT)
            .unwrap()
            .lines()
            .take(count)
            .collect::<Vec<_>>()
            .join("\n");
        let (_, retained, stderr) = run(
            &["play", "run", "-", "--format", "transcript"],
            script.as_bytes(),
        );
        assert!(stderr.is_empty());
        let evidence = fart_services::play::Transcript::from_json(&retained).unwrap();
        let expected = evidence.reconstruct().unwrap();
        let (code, stdout, stderr) =
            run(&["play", "reconstruct", "-", "--format", "json"], &retained);
        assert_eq!(code, 0);
        assert!(stderr.is_empty());
        assert_eq!(stdout, (expected.to_json() + "\n").as_bytes());
        assert_eq!(expected.is_complete(), count == 8);
        assert!(stdout.len() < 16 * 1024 * 1024);
        let (code, text, stderr) = run(&["play", "reconstruct", "-"], &retained);
        assert_eq!(code, 0);
        assert!(stderr.is_empty());
        assert_eq!(text, expected.to_text().as_bytes());
        let directory = TempDirectory::new();
        let path = directory.0.join("retained evidence.json");
        fs::write(&path, &retained).unwrap();
        let output = Command::new(env!("CARGO_BIN_EXE_fart"))
            .args(["play", "reconstruct"])
            .arg(&path)
            .args(["--format", "json"])
            .output()
            .unwrap();
        assert!(output.status.success());
        assert!(output.stderr.is_empty());
        assert_eq!(output.stdout, stdout);
        assert_eq!(fs::read(path).unwrap(), retained);
        let args = ["play", "reconstruct", "-", "--format", "json"].map(OsString::from);
        let mut stderr = Vec::new();
        assert_eq!(
            fart_cli::run(&args, &mut &retained[..], &mut FailingWriter, &mut stderr),
            1
        );
        assert!(String::from_utf8(stderr).unwrap().contains("write output"));
    }
    for format in ["text", "json"] {
        let (code, stdout, stderr) = run(&["play", "reconstruct", "-", "--format", format], b"{}");
        assert_eq!(code, 1);
        assert_eq!(stdout.is_empty(), format == "text");
        assert_eq!(stderr.is_empty(), format == "json");
    }
    let args = ["play", "reconstruct", "-", "--format", "json"].map(OsString::from);
    let mut reader = CountedReader { read: 0 };
    let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
    assert_eq!(
        fart_cli::run(&args, &mut reader, &mut stdout, &mut stderr),
        1
    );
    assert_eq!(reader.read, fart_services::play::MAX_TRANSCRIPT_BYTES + 1);
    let issue: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(issue["reason_code"], "input_too_large");
}

#[test]
fn restriction_commands_are_bounded_portable_and_preserve_full_service_reports() {
    use fart_services::restriction;
    for (operation, fixture, report) in [
        (
            "predict",
            include_bytes!("../../../testdata/restriction/gamma15-choked.json").as_slice(),
            restriction::predict as fn(&[u8]) -> restriction::Report,
        ),
        (
            "history",
            include_bytes!("../../../testdata/restriction/gamma15-choked-history.json").as_slice(),
            restriction::history as fn(&[u8]) -> restriction::Report,
        ),
    ] {
        let expected = report(fixture);
        let (code, stdout, stderr) = run(
            &["restriction", operation, "-", "--format", "json"],
            fixture,
        );
        assert_eq!(code, 0, "{}", String::from_utf8_lossy(&stdout));
        assert!(stderr.is_empty());
        assert_eq!(stdout, (expected.to_json() + "\n").as_bytes());
        let (code, text, stderr) = run(&["restriction", operation, "-"], fixture);
        assert_eq!(code, 0);
        assert!(stderr.is_empty());
        assert_eq!(text, expected.to_text().as_bytes());
        let directory = TempDirectory::new();
        let path = directory.0.join("restriction input.json");
        fs::write(&path, fixture).unwrap();
        let result = Command::new(env!("CARGO_BIN_EXE_fart"))
            .args(["restriction", operation])
            .arg(&path)
            .args(["--format", "json"])
            .output()
            .unwrap();
        assert!(result.status.success());
        assert!(result.stderr.is_empty());
        assert_eq!(result.stdout, stdout);
        assert_eq!(fs::read(path).unwrap(), fixture);
        for format in ["text", "json"] {
            let (code, stdout, stderr) =
                run(&["restriction", operation, "-", "--format", format], b"{}");
            assert_eq!(code, 1);
            assert_eq!(stdout.is_empty(), format == "text");
            assert_eq!(stderr.is_empty(), format == "json");
        }
        let (code, stdout, stderr) = run(
            &[
                "restriction",
                operation,
                "missing-restriction.json",
                "--format",
                "json",
            ],
            b"",
        );
        assert_eq!(code, 1);
        assert!(stderr.is_empty());
        let issue: Value = serde_json::from_slice(&stdout).unwrap();
        assert_eq!(issue["diagnostics"][0]["reason_code"], "input_not_found");
        let args = ["restriction", operation, "-", "--format", "json"].map(OsString::from);
        let mut reader = CountedReader { read: 0 };
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        assert_eq!(
            fart_cli::run(&args, &mut reader, &mut stdout, &mut stderr),
            1
        );
        assert_eq!(reader.read, MAX_INPUT_BYTES + 1);
        let issue: Value = serde_json::from_slice(&stdout).unwrap();
        assert_eq!(issue["diagnostics"][0]["reason_code"], "input_too_large");
        stdout.clear();
        assert_eq!(
            fart_cli::run(&args, &mut FailingReader, &mut stdout, &mut stderr),
            1
        );
        let issue: Value = serde_json::from_slice(&stdout).unwrap();
        assert_eq!(issue["diagnostics"][0]["reason_code"], "input_unavailable");
        assert_eq!(
            fart_cli::run(&args, &mut &fixture[..], &mut FailingWriter, &mut stderr),
            1
        );
        assert!(String::from_utf8(stderr).unwrap().contains("write output"));
    }
}

#[test]
fn native_reconstruction_rejects_rehashed_numerical_drift_with_fresh_evidence() {
    use serde_json::json;
    use sha2::{Digest, Sha256};
    fn digest(domain: &str, value: &Value) -> Value {
        let envelope = json!({"profile":fart_services::play::FINGERPRINT_PROFILE,"domain":domain,"value":value});
        let bytes = serde_json_canonicalizer::to_vec(&envelope).unwrap();
        let hex = Sha256::digest(bytes)
            .iter()
            .map(|byte| format!("{byte:02x}"))
            .collect::<String>();
        json!(format!("sha256:{hex}"))
    }
    let (_, original, _) = run(&["play", "run", "-", "--format", "transcript"], PLAY_SCRIPT);
    let mut forged: Value = serde_json::from_slice(&original).unwrap();
    forged["journal"][0]["receipt"]["report"]["final"]["pressure_pa"] = json!(12345.0);
    let baseline = forged["summary"]["baseline_fingerprint"].clone();
    let mut previous = forged["genesis"]["receipt"]["journal_fingerprint"].clone();
    let mut account = forged["genesis"]["receipt"]["account_fingerprint"].clone();
    for entry in forged["journal"].as_array_mut().unwrap() {
        if entry["receipt"]["report"]["status"] == "predicted" {
            account = digest(
                "account",
                &json!({"baseline_fingerprint":baseline,"report":entry["receipt"]["report"]}),
            );
        }
        entry["receipt"]["account_fingerprint"] = account.clone();
        entry["receipt"]["previous_journal_fingerprint"] = previous;
        entry["receipt"]
            .as_object_mut()
            .unwrap()
            .remove("journal_fingerprint");
        previous = digest("journal-entry", entry);
        entry["receipt"]["journal_fingerprint"] = previous.clone();
    }
    forged["summary"]["account_fingerprint"] = account;
    forged["summary"]["journal_fingerprint"] = previous;
    let bytes = serde_json::to_vec(&forged).unwrap();
    assert_eq!(
        run(&["play", "replay", "-", "--format", "json"], &bytes).0,
        0
    );
    for format in ["json", "text"] {
        let (code, stdout, stderr) = run(&["play", "reconstruct", "-", "--format", format], &bytes);
        assert_eq!(code, 1);
        assert!(stderr.is_empty());
        if format == "json" {
            let result: Value = serde_json::from_slice(&stdout).unwrap();
            assert_eq!(result["status"], "mismatched");
            assert_eq!(result["retained_summary"], forged["summary"]);
            assert_eq!(
                result["reconstructed_transcript"],
                serde_json::from_slice::<Value>(&original).unwrap()
            );
            assert_eq!(
                result["first_difference"],
                "/journal/0/receipt/report/final/pressure_pa"
            );
        } else {
            assert!(
                String::from_utf8(stdout)
                    .unwrap()
                    .contains("PLAY RECONSTRUCTION MISMATCHED")
            );
        }
    }
}

struct FragmentedReader<'a> {
    bytes: &'a [u8],
    width: usize,
}
impl Read for FragmentedReader<'_> {
    fn read(&mut self, output: &mut [u8]) -> io::Result<usize> {
        let count = self.width.min(output.len()).min(self.bytes.len());
        output[..count].copy_from_slice(&self.bytes[..count]);
        self.bytes = &self.bytes[count..];
        Ok(count)
    }
}

struct ShortWriter {
    bytes: Vec<u8>,
    flushes: usize,
    fail_flush: bool,
}
impl Write for ShortWriter {
    fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
        let count = bytes.len().min(7);
        self.bytes.extend_from_slice(&bytes[..count]);
        Ok(count)
    }
    fn flush(&mut self) -> io::Result<()> {
        self.flushes += 1;
        if self.fail_flush {
            Err(io::Error::other("flush failed"))
        } else {
            Ok(())
        }
    }
}

#[test]
fn play_framing_is_portable_and_short_writes_flush_each_receipt() {
    let args = ["play", "run", "-", "--format", "jsonl"].map(OsString::from);
    let (_, expected, _) = run(&["play", "run", "-", "--format", "jsonl"], PLAY_SCRIPT);
    for separator in ["\n", "\r\n"] {
        let script = std::str::from_utf8(PLAY_SCRIPT)
            .unwrap()
            .lines()
            .collect::<Vec<_>>()
            .join(separator);
        for width in [1, 7, 4096] {
            let mut reader = FragmentedReader {
                bytes: script.as_bytes(),
                width,
            };
            let mut writer = ShortWriter {
                bytes: Vec::new(),
                flushes: 0,
                fail_flush: false,
            };
            let mut stderr = Vec::new();
            assert_eq!(
                fart_cli::run(&args, &mut reader, &mut writer, &mut stderr),
                0
            );
            assert_eq!(writer.bytes, expected);
            assert_eq!(writer.flushes, 9);
            assert!(stderr.is_empty());
        }
    }
    let mut reader = FragmentedReader {
        bytes: PLAY_SCRIPT,
        width: 1,
    };
    let mut writer = ShortWriter {
        bytes: Vec::new(),
        flushes: 0,
        fail_flush: true,
    };
    let mut stderr = Vec::new();
    assert_eq!(
        fart_cli::run(&args, &mut reader, &mut writer, &mut stderr),
        1
    );
    assert_eq!(writer.flushes, 1);
    assert!(
        !reader.bytes.is_empty(),
        "output failure consumed later commands"
    );
    assert_eq!(
        writer.bytes.iter().filter(|byte| **byte == b'\n').count(),
        1
    );
    assert!(
        std::str::from_utf8(&stderr)
            .unwrap()
            .contains("write output")
    );
}

#[test]
fn play_transport_limits_rejections_and_read_failures_stay_visible() {
    let mut padded = vec![b'\n'];
    padded.extend_from_slice(PLAY_SCRIPT);
    let (code, output, stderr) = run(&["play", "run", "-", "--format", "jsonl"], &padded);
    assert_eq!(code, 1);
    assert!(stderr.is_empty());
    let rows: Vec<Value> = std::str::from_utf8(&output)
        .unwrap()
        .lines()
        .map(|line| serde_json::from_str(line).unwrap())
        .collect();
    assert_eq!(rows[0]["status"], "rejected");
    assert_eq!(rows.last().unwrap()["complete"], true);
    for (data, expected) in [
        (vec![b' '; 65_537], "64 KiB"),
        (b"{}\n".repeat(129), "128"),
        (
            [vec![b' '; 60_000], b"{}\n".to_vec()].concat().repeat(18),
            "1 MiB",
        ),
    ] {
        let (code, _, stderr) = run(&["play", "run", "-", "--format", "jsonl"], &data);
        assert_eq!(code, 1);
        assert!(std::str::from_utf8(&stderr).unwrap().contains(expected));
    }
    for command in ["run", "replay", "reconstruct"] {
        let args = ["play", command, "-"].map(OsString::from);
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        assert_eq!(
            fart_cli::run(&args, &mut FailingReader, &mut stdout, &mut stderr),
            1
        );
        assert!(stdout.is_empty());
        assert!(!stderr.is_empty());
    }
    let (code, stdout, _) = run(
        &["play", "replay", "-", "--format", "json"],
        &vec![b' '; 8 * 1024 * 1024 + 1],
    );
    assert_eq!(code, 1);
    let value: Value = serde_json::from_slice(&stdout).unwrap();
    assert_eq!(value["reason_code"], "input_too_large");
    // A Unicode separator is not a JSONL frame boundary.
    let (code, stdout, stderr) = run(
        &["play", "run", "-", "--format", "jsonl"],
        "{}\u{2028}{}\n".as_bytes(),
    );
    assert_eq!(code, 1);
    assert!(stderr.is_empty());
    assert_eq!(stdout.iter().filter(|byte| **byte == b'\n').count(), 2);
}

struct InterruptedReader<'a> {
    bytes: &'a [u8],
    interrupt_next: bool,
}
impl Read for InterruptedReader<'_> {
    fn read(&mut self, output: &mut [u8]) -> io::Result<usize> {
        if self.interrupt_next {
            self.interrupt_next = false;
            return Err(io::ErrorKind::Interrupted.into());
        }
        let count = output.len().min(127).min(self.bytes.len());
        output[..count].copy_from_slice(&self.bytes[..count]);
        self.bytes = &self.bytes[count..];
        self.interrupt_next = true;
        Ok(count)
    }
}

#[test]
fn play_transient_read_interruptions_preserve_the_complete_literal_trace() {
    let args = ["play", "run", "-", "--format", "jsonl"].map(OsString::from);
    let (_, expected, _) = run(&["play", "run", "-", "--format", "jsonl"], PLAY_SCRIPT);
    let mut reader = InterruptedReader {
        bytes: PLAY_SCRIPT,
        interrupt_next: true,
    };
    let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
    assert_eq!(
        fart_cli::run(&args, &mut reader, &mut stdout, &mut stderr),
        0
    );
    assert_eq!(stdout, expected);
    assert!(stderr.is_empty());
}

struct InterruptedFlushWriter {
    bytes: Vec<u8>,
    flushes: usize,
}
impl Write for InterruptedFlushWriter {
    fn write(&mut self, bytes: &[u8]) -> io::Result<usize> {
        self.bytes.extend_from_slice(bytes);
        Ok(bytes.len())
    }
    fn flush(&mut self) -> io::Result<()> {
        self.flushes += 1;
        if self.flushes == 1 {
            Err(io::ErrorKind::Interrupted.into())
        } else {
            Ok(())
        }
    }
}

#[test]
fn transient_flush_interruptions_retry_only_the_flush_without_duplicate_output() {
    for (args, input) in [
        (vec!["3"], b"".as_slice()),
        (vec!["play", "run", "-", "--format", "jsonl"], PLAY_SCRIPT),
    ] {
        let (_, expected, _) = run(&args, input);
        let arguments: Vec<_> = args.iter().map(OsString::from).collect();
        let mut writer = InterruptedFlushWriter {
            bytes: Vec::new(),
            flushes: 0,
        };
        let mut stderr = Vec::new();
        assert_eq!(
            fart_cli::run(&arguments, &mut &input[..], &mut writer, &mut stderr),
            0
        );
        assert_eq!(writer.bytes, expected);
        assert!(writer.flushes >= 2);
        assert!(stderr.is_empty());
    }
}

#[test]
fn play_command_limit_excludes_cr_only_when_it_is_part_of_crlf_framing() {
    let mut command = PLAY_SCRIPT
        .split(|byte| *byte == b'\n')
        .next()
        .unwrap()
        .to_vec();
    command.resize(fart_services::play::MAX_COMMAND_BYTES, b' ');
    let mut crlf = command.clone();
    crlf.extend_from_slice(b"\r\n");
    let (_, stdout, stderr) = run(&["play", "run", "-", "--format", "jsonl"], &crlf);
    assert!(stderr.is_empty());
    let first: Value =
        serde_json::from_slice(stdout.split(|byte| *byte == b'\n').next().unwrap()).unwrap();
    assert_eq!(first["status"], "accepted");
    command.push(b'\r');
    let (code, stdout, stderr) = run(&["play", "run", "-", "--format", "jsonl"], &command);
    assert_eq!(code, 1);
    assert!(stdout.is_empty(), "oversized EOF payload was admitted");
    assert!(std::str::from_utf8(&stderr).unwrap().contains("64 KiB"));
}
