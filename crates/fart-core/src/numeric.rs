// Safe binary64 scaling for positive model terms. Keeping the exponent separate
// prevents a tiny intermediate product from destroying a finite final result.
// Summing equally scaled products also preserves contributions whose individual
// unscaled products would underflow before normalization by the total mass.

pub(super) fn product_over(values: &[f64], divisor: f64) -> f64 {
    let (fraction, exponent) = product_parts(values);
    let (divisor_fraction, divisor_exponent) = split(divisor);
    scale(fraction / divisor_fraction, exponent - divisor_exponent)
}

pub(super) fn sum_products_over<const N: usize>(
    values: impl Iterator<Item = [f64; N]>,
    divisor: f64,
) -> f64 {
    let terms: Vec<_> = values.map(|term| product_parts(&term)).collect();
    let exponent = terms.iter().map(|term| term.1).max().unwrap_or(0);
    let fraction = super::stable_sum(
        terms
            .into_iter()
            .map(|term| scale(term.0, term.1 - exponent))
            .collect(),
    );
    let (divisor_fraction, divisor_exponent) = split(divisor);
    scale(fraction / divisor_fraction, exponent - divisor_exponent)
}

fn product_parts(values: &[f64]) -> (f64, i32) {
    let (mut fraction, mut exponent) = (1.0, 0);
    for &value in values {
        let (part, power) = split(value);
        let (normalized, normalization_power) = split(fraction * part);
        fraction = normalized;
        exponent += power + normalization_power;
    }
    (fraction, exponent)
}

// Equivalent to frexp for nonnegative values, including exact subnormal inputs.
fn split(value: f64) -> (f64, i32) {
    if value == 0.0 || !value.is_finite() {
        return (value, 0);
    }
    let bits = value.to_bits();
    let exponent = ((bits >> 52) & 0x7ff) as i32;
    if exponent == 0 {
        let (fraction, power) = split(value * (1_u64 << 54) as f64);
        return (fraction, power - 54);
    }
    let fraction = f64::from_bits((bits & ((1_u64 << 52) - 1)) | (1022_u64 << 52));
    (fraction, exponent - 1022)
}

// Round once at the final subnormal boundary. Direct powi(exponent) would itself
// underflow or overflow at the edges even when fraction*2^exponent is finite.
fn scale(fraction: f64, exponent: i32) -> f64 {
    if fraction == 0.0 || !fraction.is_finite() {
        return fraction;
    }
    let (fraction, normalization_power) = split(fraction);
    let exponent = exponent + normalization_power;
    if exponent > 1024 {
        return f64::INFINITY;
    }
    if exponent < -1074 {
        return 0.0;
    }
    if exponent == 1024 {
        return (fraction * f64::from_bits(2046_u64 << 52)) * 2.0;
    }
    if exponent < -1021 {
        return (fraction * 2.0_f64.powi(exponent + 1022)) * f64::MIN_POSITIVE;
    }
    fraction * f64::from_bits(((exponent + 1023) as u64) << 52)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn product_scaling_has_exact_binary64_boundary_anchors() {
        for value in [
            0.0,
            f64::from_bits(1),
            f64::MIN_POSITIVE,
            0.5,
            1.0,
            f64::MAX,
        ] {
            assert_eq!(product_over(&[value], 1.0), value);
        }
        assert_eq!(product_over(&[f64::MAX, 2.0], 2.0), f64::MAX);
        assert_eq!(
            product_over(&[f64::from_bits(1), 0.5], 0.5),
            f64::from_bits(1)
        );
        assert_eq!(product_over(&[f64::from_bits(1), 0.5], 1.0), 0.0);
        assert_eq!(
            product_over(&[f64::from_bits(1), 0.75], 1.0),
            f64::from_bits(1)
        );
        assert_eq!(product_over(&[f64::from_bits(1), 0.125], 1.0), 0.0);
        assert_eq!(product_over(&[f64::MAX, 4.0], 1.0), f64::INFINITY);
        assert_eq!(product_over(&[f64::INFINITY], 1.0), f64::INFINITY);
        assert_eq!(
            sum_products_over([[f64::from_bits(1), 0.5]; 2].into_iter(), 1.0),
            f64::from_bits(1)
        );
    }
}
