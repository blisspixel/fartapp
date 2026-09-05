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
    ] {
        let arguments: Vec<_> = args.iter().map(OsString::from).collect();
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        let mut reader = FailingReader;
        assert_eq!(
            fart_cli::run(&arguments, &mut reader, &mut stdout, &mut stderr),
            0
        );
        let text = String::from_utf8(stdout).unwrap();
        assert!(text.contains("experimental native Rust CLI"));
        assert!(text.contains("no PlayService"));
        assert!(stderr.is_empty());
    }
}

#[test]
fn usage_errors_do_not_produce_success_output() {
    for args in [
        vec![],
        vec!["1", "2"],
        vec!["help", "play"],
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
