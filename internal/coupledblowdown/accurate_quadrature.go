package coupledblowdown

import (
	"math"

	"github.com/blisspixel/fartapp/internal/floatmath"
)

// The embedded 7/15 Gauss-Kronrod nodes, weights, and error rescaling follow
// the published QUADPACK DQK15 rule: https://www.netlib.org/quadpack/dqk15.f
// This is an a posteriori estimator, not an interval-arithmetic certificate.
var accurateNodes = [...]float64{
	0.99145537112081263921, 0.94910791234275852453, 0.86486442335976907279,
	0.74153118559939443986, 0.58608723546769113029, 0.40584515137739716691,
	0.20778495500789846760, 0,
}

var accurateKronrodWeights = [...]float64{
	0.022935322010529224964, 0.063092092629978553291, 0.10479001032225018384,
	0.14065325971552591875, 0.16900472663926790283, 0.19035057806478540991,
	0.20443294007529889241, 0.20948214108472782801,
}

var accurateGaussWeights = [...]float64{
	0.12948496616886969327, 0.27970539148927666790, 0.38183005050511894495, 0.41795918367346938776,
}

type accurateQuadrature struct {
	path     accuratePath
	options  AccuracyOptions
	evidence *AccuracyEvidence
}

func (q *accurateQuadrature) integrate(lower, upper float64) ([3]float64, [3]float64, error) {
	var total, errors [3]accumulator
	start := lower
	for _, cut := range append(append([]float64(nil), q.path.breaks...), upper) {
		if cut <= start || cut > upper {
			continue
		}
		value, estimate, err := q.adaptive(start, cut, 0)
		if err != nil {
			return [3]float64{}, [3]float64{}, err
		}
		for index := range value {
			total[index].add(value[index])
			errors[index].add(estimate[index])
		}
		start = cut
	}
	return [3]float64{total[0].value(), total[1].value(), total[2].value()},
		[3]float64{errors[0].value(), errors[1].value(), errors[2].value()}, nil
}

func (q *accurateQuadrature) adaptive(lower, upper float64, depth int) ([3]float64, [3]float64, error) {
	value, estimate, err := q.rule(lower, upper)
	if err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	accepted := true
	for index := range value {
		tolerance := q.options.RelativeTolerance * value[index]
		if index == 0 {
			tolerance += q.options.AbsoluteTimeToleranceSeconds * (upper - lower)
		}
		if !finite(value[index]) || !finite(estimate[index]) || !finite(tolerance) {
			return [3]float64{}, [3]float64{}, ErrAccuracyNotAchieved
		}
		if value[index] > 0 && tolerance < math.SmallestNonzeroFloat64 {
			return [3]float64{}, [3]float64{}, ErrAccuracyNotAchieved
		}
		accepted = accepted && estimate[index] <= tolerance
	}
	if accepted {
		q.evidence.AcceptedIntervals++
		return value, estimate, nil
	}
	middle := lower + (upper-lower)/2
	if depth >= 48 || middle <= lower || middle >= upper {
		return [3]float64{}, [3]float64{}, ErrAccuracyNotAchieved
	}
	q.evidence.Refinements++
	left, leftError, err := q.adaptive(lower, middle, depth+1)
	if err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	right, rightError, err := q.adaptive(middle, upper, depth+1)
	if err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	for index := range left {
		left[index] += right[index]
		leftError[index] += rightError[index]
	}
	return left, leftError, nil
}

func (q *accurateQuadrature) rule(lower, upper float64) ([3]float64, [3]float64, error) {
	if q.evidence.Evaluations > q.options.MaxEvaluations-15 {
		return [3]float64{}, [3]float64{}, ErrAccuracyBudgetExhausted
	}
	half := (upper - lower) / 2
	middle := lower + half
	var samples [15][3]float64
	var kronrod, gauss [3]accumulator
	for node, offset := range accurateNodes {
		for side := 0; side < 2; side++ {
			if node == 7 && side == 1 {
				continue
			}
			z := middle - half*offset
			if side == 1 {
				z = middle + half*offset
			}
			q.evidence.Evaluations++
			value, err := q.path.integrand(z)
			if err != nil {
				return [3]float64{}, [3]float64{}, err
			}
			for index := range value {
				// Scale by interval width before accumulation so finite
				// integrals do not overflow solely from large point rates.
				scaled := floatmath.Product(half, value[index])
				samples[2*node+side][index] = scaled
				kronrod[index].add(accurateKronrodWeights[node] * scaled)
				if node%2 == 1 {
					gauss[index].add(accurateGaussWeights[node/2] * scaled)
				}
			}
		}
	}
	var value, estimate [3]float64
	for index := range value {
		value[index] = kronrod[index].value()
		var deviation accumulator
		for node := range accurateNodes {
			deviation.add(accurateKronrodWeights[node] * math.Abs(samples[2*node][index]-value[index]/2))
			if node < 7 {
				deviation.add(accurateKronrodWeights[node] * math.Abs(samples[2*node+1][index]-value[index]/2))
			}
		}
		difference := math.Abs(value[index] - gauss[index].value())
		if deviation.value() > 0 && difference > 0 {
			difference = deviation.value() * math.Pow(math.Min(1, 200*(difference/deviation.value())), 1.5)
		}
		estimate[index] = math.Max(difference, 50*(math.Nextafter(1, 2)-1)*value[index])
		if value[index] > 0 {
			estimate[index] = math.Max(estimate[index], math.SmallestNonzeroFloat64)
		}
	}
	return value, estimate, nil
}
