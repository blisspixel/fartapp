use std::ffi::OsString;
use std::fs::File;
use std::io::{BufRead, BufReader, Read, Write};

use fart_services::play::{MAX_COMMAND_BYTES, MAX_TRANSCRIPT_BYTES, PlayService, Transcript};

use crate::{Format, diagnostic, help, options, output};

const MAX_STREAM_BYTES: usize = 1024 * 1024;
const MAX_COMMANDS: usize = 128;
const MAX_OUTPUT_BYTES: usize = 16 * 1024 * 1024;

pub(super) fn run(
    args: &[OsString],
    stdin: &mut dyn Read,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> u8 {
    if args.len() == 1 && matches!(args[0].to_str(), Some("help" | "-h" | "--help")) {
        return output(help::PLAY.as_bytes(), stdout, stderr);
    }
    let Some(operation @ ("run" | "replay")) = args.first().and_then(|arg| arg.to_str()) else {
        return diagnostic(
            stderr,
            "usage: fart play <run|replay>; use 'fart help play'\n",
        );
    };
    let options = match options(&args[1..], operation == "run") {
        Ok(options) => options,
        Err(message) => return diagnostic(stderr, message),
    };
    if options.help {
        return output(
            if operation == "run" {
                help::PLAY_RUN
            } else {
                help::PLAY_REPLAY
            }
            .as_bytes(),
            stdout,
            stderr,
        );
    }
    let Some(source) = options.source else {
        return diagnostic(stderr, "missing play input; use a file or - for stdin\n");
    };
    let execute = |reader: &mut dyn Read, stdout: &mut dyn Write, stderr: &mut dyn Write| {
        if operation == "run" {
            stream(reader, options.format, stdout, stderr)
        } else {
            replay(reader, options.format, stdout, stderr)
        }
    };
    if source == "-" {
        return execute(stdin, stdout, stderr);
    }
    match File::open(source) {
        Ok(mut file) => execute(&mut file, stdout, stderr),
        Err(_) => diagnostic(
            stderr,
            "play input unavailable; check the path and read access, or use - for stdin\n",
        ),
    }
}

fn stream(
    reader: &mut dyn Read,
    format: Format,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> u8 {
    // The outer take also bounds buffered read-ahead. Framing is LF only, so
    // Unicode line/paragraph separators remain data for the strict JSON parser.
    let mut reader = BufReader::new(reader.take((MAX_STREAM_BYTES + 1) as u64));
    let mut service = PlayService::new();
    let (mut input_bytes, mut output_bytes, mut count, mut rejected) = (0, 0, 0, false);
    loop {
        match reader.fill_buf() {
            Ok([]) => break,
            Ok(_) if count == MAX_COMMANDS => {
                return diagnostic(
                    stderr,
                    "play delivery incomplete: command limit exceeded (128); no finish was inferred\n",
                );
            }
            Err(error) if error.kind() == std::io::ErrorKind::Interrupted => continue,
            Err(_) => {
                return diagnostic(
                    stderr,
                    "play delivery incomplete: input read failed; no finish was inferred\n",
                );
            }
            _ => {}
        }
        let command = match next_command(&mut reader, &mut input_bytes) {
            Ok(command) => command,
            Err(message) => return diagnostic(stderr, message),
        };
        count += 1;
        let reply = service.process_json(&command);
        rejected |= reply.is_rejected();
        let text = match format {
            Format::Text => reply.to_text() + "\n",
            Format::Jsonl => reply.to_json() + "\n",
            Format::Transcript => continue,
            Format::Json => unreachable!("run format validated"),
        };
        if emit(&text, &mut output_bytes, stdout, stderr) != 0 {
            return 1;
        }
    }
    let summary = service.end_of_input();
    let text = match format {
        Format::Text => summary.to_text(),
        Format::Jsonl => summary.to_json() + "\n",
        Format::Transcript => match summary.transcript() {
            Some(transcript) => transcript.to_json() + "\n",
            None => {
                return diagnostic(
                    stderr,
                    "no transcript retained: session was never started\n",
                );
            }
        },
        Format::Json => unreachable!("run format validated"),
    };
    if emit(&text, &mut output_bytes, stdout, stderr) != 0 || !summary.is_complete() || rejected {
        1
    } else {
        0
    }
}

fn next_command(reader: &mut impl BufRead, total: &mut usize) -> Result<Vec<u8>, &'static str> {
    let mut line = Vec::new();
    reader
        .take((MAX_COMMAND_BYTES + 2) as u64)
        .read_until(b'\n', &mut line)
        .map_err(|_| "play delivery incomplete: input read failed; no finish was inferred\n")?;
    *total += line.len();
    if *total > MAX_STREAM_BYTES {
        return Err(
            "play delivery incomplete: input limit exceeded (1 MiB); no finish was inferred\n",
        );
    }
    if line.last() == Some(&b'\n') {
        line.pop();
        if line.last() == Some(&b'\r') {
            line.pop();
        }
    }
    if line.len() > MAX_COMMAND_BYTES {
        return Err(
            "play delivery incomplete: command limit exceeded (64 KiB); no finish was inferred\n",
        );
    }
    Ok(line)
}

fn emit(text: &str, total: &mut usize, stdout: &mut dyn Write, stderr: &mut dyn Write) -> u8 {
    if text.len() > MAX_OUTPUT_BYTES - *total {
        return diagnostic(
            stderr,
            "play delivery incomplete: output limit exceeded (16 MiB); inspect the last delivered receipt\n",
        );
    }
    *total += text.len();
    output(text.as_bytes(), stdout, stderr)
}

fn replay(
    reader: &mut dyn Read,
    format: Format,
    stdout: &mut dyn Write,
    stderr: &mut dyn Write,
) -> u8 {
    let mut data = Vec::new();
    if reader
        .take((MAX_TRANSCRIPT_BYTES + 1) as u64)
        .read_to_end(&mut data)
        .is_err()
    {
        return diagnostic(
            stderr,
            "play replay input unavailable; provide a complete transcript file or stream\n",
        );
    }
    match Transcript::from_json(&data).and_then(|transcript| transcript.replay()) {
        Ok(summary) => output(
            if matches!(format, Format::Text) {
                summary.to_text()
            } else {
                summary.to_json() + "\n"
            }
            .as_bytes(),
            stdout,
            stderr,
        ),
        Err(issue) => {
            if matches!(format, Format::Json) {
                let _ = output((issue.to_json() + "\n").as_bytes(), stdout, stderr);
                1
            } else {
                diagnostic(
                    stderr,
                    &format!(
                        "PLAY REPLAY REFUSED\n\nReason: {}\nPath: {:?}\nRecovery: provide the exact retained transcript from play run --format transcript.\n",
                        issue.reason(),
                        issue.path().chars().take(256).collect::<String>()
                    ),
                )
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn aggregate_output_limit_does_not_write_a_partial_record() {
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        let mut used = MAX_OUTPUT_BYTES - 3;
        assert_eq!(emit("abc", &mut used, &mut stdout, &mut stderr), 0);
        assert_eq!(emit("d", &mut used, &mut stdout, &mut stderr), 1);
        assert_eq!(stdout, b"abc");
        assert_eq!(used, MAX_OUTPUT_BYTES);
        assert!(String::from_utf8(stderr).unwrap().contains("output limit"));
    }
}
