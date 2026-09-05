use serde_json::{Value, json};
use sha2::{Digest, Sha256};

use super::FINGERPRINT_PROFILE;

pub(super) fn canonical(value: &Value) -> Vec<u8> {
    // Values come only from the strict parser or owned finite, string-keyed
    // constructors. Bounded integer control fields are exactly binary64-safe.
    serde_json_canonicalizer::to_vec(value).expect("validated JSON value is canonicalizable")
}

pub(super) fn digest(domain: &str, value: &Value) -> String {
    let envelope = json!({"profile":FINGERPRINT_PROFILE,"domain":domain,"value":value});
    let hash = Sha256::digest(canonical(&envelope));
    let mut result = String::with_capacity(71);
    result.push_str("sha256:");
    for byte in hash {
        use std::fmt::Write;
        write!(&mut result, "{byte:02x}").expect("string write");
    }
    result
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rfc8785_number_and_utf16_sorting_vectors() {
        let value: Value = serde_json::from_str(r#"{"numbers":[333333333.33333329,1E30,4.50,2e-3,0.000000000000000000000000001],"\uE000":1,"\uD83D\uDE00":2,"negative_zero":-0.0}"#).unwrap();
        assert_eq!(
            String::from_utf8(canonical(&value)).unwrap(),
            "{\"negative_zero\":0,\"numbers\":[333333333.3333333,1e+30,4.5,0.002,1e-27],\"\u{1f600}\":2,\"\u{e000}\":1}"
        );
    }

    #[test]
    fn sha256_known_answer_and_domain_separation() {
        let actual = Sha256::digest(b"abc")
            .iter()
            .map(|byte| format!("{byte:02x}"))
            .collect::<String>();
        assert_eq!(
            actual,
            "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
        );
        assert_ne!(
            digest("account", &json!({"a":1})),
            digest("request", &json!({"a":1}))
        );
    }
}
