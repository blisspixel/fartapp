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

#[cfg(test)]
mod tests {
    use super::human_number;

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
