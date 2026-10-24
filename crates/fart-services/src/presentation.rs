//! Human-only formatting shared by native report and session views.

/// Six significant digits with notation chosen after rounding. This does not
/// change retained values, tolerances, identities, or JSON serialization.
pub fn human_number(value: f64) -> String {
    if !value.is_finite() {
        return "unavailable".into();
    }
    if value == 0.0 {
        return "0".into();
    }
    let rounded = format!("{value:.5e}");
    let (mantissa, exponent_text) = rounded.split_once('e').expect("finite scientific format");
    let exponent: i32 = exponent_text.parse().expect("formatted decimal exponent");
    if !(-3..6).contains(&exponent) {
        return format!(
            "{}e{exponent}",
            mantissa.trim_end_matches('0').trim_end_matches('.')
        );
    }
    let (sign, magnitude) = mantissa
        .strip_prefix('-')
        .map_or(("", mantissa), |v| ("-", v));
    let digits = magnitude.replace('.', "");
    let point = exponent + 1;
    let decimal = if point <= 0 {
        format!("0.{}{}", "0".repeat((-point) as usize), digits)
    } else {
        let (whole, fraction) = digits.split_at(point as usize);
        format!("{whole}.{fraction}")
    };
    format!(
        "{sign}{}",
        decimal.trim_end_matches('0').trim_end_matches('.')
    )
}

/// Wrap whitespace-separated words at limit columns with a hanging indent.
/// Unbroken tokens exceeding limit are never wrapped mid-token.
pub fn wrap_prose(text: &str, indent: &str, limit: usize) -> String {
    let words: Vec<&str> = text.split_whitespace().collect();
    if words.is_empty() {
        return String::new();
    }
    let mut lines = Vec::new();
    let mut current_line = format!("{indent}{}", words[0]);
    for &word in &words[1..] {
        if current_line.len() + 1 + word.len() <= limit {
            current_line.push(' ');
            current_line.push_str(word);
        } else {
            lines.push(current_line);
            current_line = format!("{indent}{word}");
        }
    }
    lines.push(current_line);
    lines.join("\n")
}

#[cfg(test)]
mod tests {
    use super::{human_number, wrap_prose};

    #[test]
    fn wrap_prose_wraps_at_limit_with_indent() {
        let text = "This is a long sentence that should be wrapped across multiple lines when the column limit is reached.";
        let wrapped = wrap_prose(text, "  ", 40);
        for line in wrapped.lines() {
            assert!(line.len() <= 40);
            assert!(line.starts_with("  "));
        }
    }

    #[test]
    fn wrap_prose_preserves_unbroken_token_over_limit() {
        let text = "word supercalifragilisticexpialidocious_unbroken_token_that_exceeds_the_limit end";
        let wrapped = wrap_prose(text, "  ", 30);
        let lines: Vec<_> = wrapped.lines().collect();
        assert_eq!(lines.len(), 3);
        assert_eq!(lines[0], "  word");
        assert_eq!(lines[1], "  supercalifragilisticexpialidocious_unbroken_token_that_exceeds_the_limit");
        assert_eq!(lines[2], "  end");
    }

    #[test]
    fn shared_presentation_cases() {
        let cases: serde_json::Value =
            serde_json::from_str(include_str!("../../../testdata/presentation/numbers.json"))
                .unwrap();
        for item in cases.as_array().unwrap() {
            let value: f64 = item["input"].as_str().unwrap().parse().unwrap();
            assert_eq!(human_number(value), item["display"].as_str().unwrap());
            if value != 0.0 {
                assert_ne!(human_number(value), "0");
            }
        }
        for value in [f64::NAN, f64::INFINITY, f64::NEG_INFINITY] {
            assert_eq!(human_number(value), "unavailable");
        }
    }
}
