//! Thin native command presentation over the stateless service facade.

use std::ffi::{OsStr, OsString};
use std::fs::File;
use std::io::{self, Read, Write};

use fart_services::{
    InputFailure, MAX_INPUT_BYTES, PredictionReport, intensity_reply, predict_reservoir,
    reservoir_input_failure,
};

const HELP: &str = "F.A.R.T. Lab experimental native Rust CLI\n\nUsage:\n  fart <intensity>\n  fart reservoir predict <request.json|-> [--format text|json]\n  fart help reservoir predict\n\nIntensity is the permanent toy example, an integer from 1 to 5.\nThe reservoir command is an analytical SI endpoint candidate. It reads at most\n65,536 bytes and requires explicit model, closure, units, and component inputs.\nJSON input rejects duplicate, unknown, and null members.\n\nUse - for standard input. Text is the default; --format json emits one report.\nExit status 0 means completion; 1 means usage, input, model, or output failure.\nA refused JSON request still emits its structured report.\n\nThis build provides no PlayService, case archive, protocol server, empirical\nvalidation, plume, physical audio, or general law-capability contract.\n";

#[derive(Clone, Copy)]
enum Format {
    Text,
    Json,
}

struct Options<'a> {
    source: Option<&'a OsStr>,
    format: Format,
    help: bool,
}

/// Execute arguments excluding the program name using injectable standard streams.
/// File reads are bounded; commands never write or replace an input file.
pub fn run(
    args: &[OsString],
    stdin: &mut dyn Read,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> u8 {
    if args.len() == 1 && matches!(args[0].to_str(), Some("-h" | "--help")) {
        return output(HELP.as_bytes(), stdout, stderr);
    }
    if args.first().is_some_and(|arg| arg == "help") {
        let valid = args.len() == 1
            || args.len() == 2 && args[1] == "reservoir"
            || args.len() == 3 && args[1] == "reservoir" && args[2] == "predict";
        return if valid {
            output(HELP.as_bytes(), stdout, stderr)
        } else {
            diagnostic(stderr, "unknown help topic\n")
        };
    }
    if args.first().is_some_and(|arg| arg == "reservoir") {
        if args.len() == 2 && matches!(args[1].to_str(), Some("-h" | "--help")) {
            return output(HELP.as_bytes(), stdout, stderr);
        }
        if args.get(1).is_none_or(|arg| arg != "predict") {
            return diagnostic(
                stderr,
                "usage: fart reservoir predict <request.json|-> [--format text|json]\n",
            );
        }
        let options = match options(&args[2..]) {
            Ok(options) => options,
            Err(message) => return diagnostic(stderr, message),
        };
        if options.help {
            return output(HELP.as_bytes(), stdout, stderr);
        }
        let Some(source) = options.source else {
            return diagnostic(
                stderr,
                "missing reservoir input; use a JSON file or - for stdin\n",
            );
        };
        let report = read_prediction(source, stdin);
        let success = report.is_predicted();
        let text = match options.format {
            Format::Text => report.to_text(),
            Format::Json => report.to_json() + "\n",
        };
        if !success && matches!(options.format, Format::Text) {
            return diagnostic(stderr, &text);
        }
        return if output(text.as_bytes(), stdout, stderr) == 0 && success {
            0
        } else {
            1
        };
    }
    if args.len() != 1 {
        return diagnostic(stderr, "usage: fart <intensity>\n");
    }
    if let Some(reply) = args[0].to_str().and_then(intensity_reply) {
        return output(reply.as_bytes(), stdout, stderr);
    }
    let token: String = args[0].to_string_lossy().chars().take(32).collect();
    diagnostic(
        stderr,
        &format!("invalid intensity {token:?}: must be a canonical integer from 1 to 5\n"),
    )
}

fn options(args: &[OsString]) -> Result<Options<'_>, &'static str> {
    let mut result = Options {
        source: None,
        format: Format::Text,
        help: false,
    };
    let mut format_seen = false;
    let mut flags = true;
    let mut index = 0;
    while index < args.len() {
        let arg = &args[index];
        if flags && arg == "--" {
            flags = false;
        } else if flags && matches!(arg.to_str(), Some("-h" | "--help")) {
            if result.help {
                return Err("--help may be specified only once\n");
            }
            result.help = true;
        } else if flags && arg == "--format" {
            if format_seen {
                return Err("--format may be specified only once\n");
            }
            format_seen = true;
            index += 1;
            result.format = match args.get(index).and_then(|value| value.to_str()) {
                Some("text") => Format::Text,
                Some("json") => Format::Json,
                _ => return Err("--format requires text or json\n"),
            };
        } else if flags
            && arg
                .to_str()
                .is_some_and(|value| value.starts_with('-') && value != "-")
        {
            return Err("unknown reservoir option; use -- before a filename beginning with -\n");
        } else if result.source.replace(arg).is_some() {
            return Err("provide exactly one reservoir input source\n");
        }
        index += 1;
    }
    Ok(result)
}

fn read_prediction(source: &OsStr, stdin: &mut dyn Read) -> PredictionReport {
    if source == "-" {
        return read_bounded(stdin);
    }
    match File::open(source) {
        Ok(mut file) => read_bounded(&mut file),
        Err(error) => {
            let failure = match error.kind() {
                io::ErrorKind::NotFound => InputFailure::NotFound,
                io::ErrorKind::PermissionDenied => InputFailure::PermissionDenied,
                _ => InputFailure::Unavailable,
            };
            reservoir_input_failure(failure, false)
        }
    }
}

fn read_bounded(reader: &mut dyn Read) -> PredictionReport {
    let mut bytes = Vec::new();
    if reader
        .take((MAX_INPUT_BYTES + 1) as u64)
        .read_to_end(&mut bytes)
        .is_err()
    {
        return reservoir_input_failure(InputFailure::Unavailable, true);
    }
    if bytes.len() > MAX_INPUT_BYTES {
        return reservoir_input_failure(InputFailure::TooLarge, true);
    }
    predict_reservoir(&bytes)
}

fn output(bytes: &[u8], stdout: &mut dyn Write, stderr: &mut dyn Write) -> u8 {
    if stdout.write_all(bytes).is_err() {
        return diagnostic(stderr, "write output: unavailable\n");
    }
    0
}

fn diagnostic(stderr: &mut dyn Write, message: &str) -> u8 {
    let _ = stderr.write_all(message.as_bytes());
    1
}
