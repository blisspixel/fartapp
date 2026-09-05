use std::ffi::{OsStr, OsString};
use std::fs::File;
use std::io::{self, Read, Write};

use fart_services::{
    InputFailure,
    restriction::{self, Kind, MAX_INPUT_BYTES, Report},
};

use crate::{Format, diagnostic, help, options, output};

pub(super) fn run(
    args: &[OsString],
    stdin: &mut dyn Read,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> u8 {
    if args.len() == 1 && matches!(args[0].to_str(), Some("help" | "-h" | "--help")) {
        return output(help::RESTRICTION.as_bytes(), stdout, stderr);
    }
    let (kind, help) = match args.first().and_then(|arg| arg.to_str()) {
        Some("predict") => (Kind::Prediction, help::RESTRICTION_PREDICT),
        Some("history") => (Kind::History, help::RESTRICTION_HISTORY),
        _ => {
            return diagnostic(
                stderr,
                "usage: fart restriction <predict|history>; use 'fart help restriction'\n",
            );
        }
    };
    let options = match options(&args[1..], false) {
        Ok(options) => options,
        Err(message) => return diagnostic(stderr, message),
    };
    if options.help {
        return output(help.as_bytes(), stdout, stderr);
    }
    let Some(source) = options.source else {
        return diagnostic(
            stderr,
            "missing restriction input; use a JSON file or - for stdin\n",
        );
    };
    let report = read(kind, source, stdin);
    let text = if matches!(options.format, Format::Text) {
        report.to_text()
    } else {
        report.to_json() + "\n"
    };
    if !report.is_predicted() && matches!(options.format, Format::Text) {
        return diagnostic(stderr, &text);
    }
    if output(text.as_bytes(), stdout, stderr) == 0 && report.is_predicted() {
        0
    } else {
        1
    }
}

fn read(kind: Kind, source: &OsStr, stdin: &mut dyn Read) -> Report {
    if source == "-" {
        return read_bounded(kind, stdin);
    }
    match File::open(source) {
        Ok(mut file) => read_bounded(kind, &mut file),
        Err(error) => {
            let failure = match error.kind() {
                io::ErrorKind::NotFound => InputFailure::NotFound,
                io::ErrorKind::PermissionDenied => InputFailure::PermissionDenied,
                _ => InputFailure::Unavailable,
            };
            restriction::input_failure(kind, failure, false)
        }
    }
}

fn read_bounded(kind: Kind, reader: &mut dyn Read) -> Report {
    let mut bytes = Vec::new();
    if reader
        .take((MAX_INPUT_BYTES + 1) as u64)
        .read_to_end(&mut bytes)
        .is_err()
    {
        return restriction::input_failure(kind, InputFailure::Unavailable, true);
    }
    if bytes.len() > MAX_INPUT_BYTES {
        return restriction::input_failure(kind, InputFailure::TooLarge, true);
    }
    match kind {
        Kind::Prediction => restriction::predict(&bytes),
        Kind::History => restriction::history(&bytes),
    }
}
