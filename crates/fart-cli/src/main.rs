//! Process entry point for the experimental native CLI.

use std::io;
use std::process::ExitCode;

fn main() -> ExitCode {
    let args = std::env::args_os().skip(1).collect::<Vec<_>>();
    ExitCode::from(fart_cli::run(
        &args,
        &mut io::stdin().lock(),
        &mut io::stdout().lock(),
        &mut io::stderr().lock(),
    ))
}
