use std::ffi::OsStr;

struct Row {
    command: &'static str,
    home: &'static str,
    example: &'static str,
}

fn rows() -> Vec<Row> {
    include_str!("../../../internal/cli/parity.tsv")
        .lines()
        .filter_map(|line| {
            let line = line.trim();
            if line.is_empty() || line.starts_with('#') {
                return None;
            }
            let mut parts = line.split('\t');
            Some(Row {
                command: parts.next()?,
                home: parts.next()?,
                example: parts.next()?,
            })
        })
        .collect()
}

pub(super) fn go_oracle_hint(command: &OsStr) -> Option<String> {
    let command = command.to_str()?;
    rows()
        .into_iter()
        .find(|row| row.command == command && row.home == "go")
        .map(|row| {
            format!(
                "'{}' is available in the Go oracle CLI: {}\n",
                row.command, row.example
            )
        })
}

pub(super) fn looks_like_intensity(token: &str) -> bool {
    token.chars().any(char::is_numeric)
}
