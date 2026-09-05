package cli

import (
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/blisspixel/fartapp/internal/coupledblowdown"
	"github.com/blisspixel/fartapp/internal/walkcase"
)

const refineHelp = `Integrate a walk case with estimated numerical accuracy.

Usage:
  fartapp walk refine <case.json|-> --relative-tolerance <value> --max-evaluations <count> [--absolute-time-tolerance <seconds>] [--format text|json]

Required:
  --relative-tolerance  Finite relative quadrature tolerance in [1e-12, 0.1].
  --max-evaluations     Integer function-evaluation budget in [15, 1000000].

Optional:
  --absolute-time-tolerance  Nonnegative seconds, default 0.
  --format text|json        English summary or complete report, default text.
  -h, --help                Show help without reading a file or standard input.

Time uses absolute tolerance + relative tolerance * elapsed time. Impulse and
stroke use the relative tolerance. Estimates are not rigorous error bounds;
they exclude input uncertainty, floating-point representation, and model error.

For flowing cases, step.max_time_s must be 0 and the resting area positive.
Both thermal closures and capped linear compliance are supported. A zero-rest
compliant opening approaches equalization at infinite time and is refused.
step.max_withdrawal_fraction_per_step selects retained samples; step.max_steps
can truncate an otherwise accurate path. Completion is reported separately.

This operation has its own implementation revision. Legacy simulate, witness,
and retained evidence continue to use their existing numerical profile.

Example:
  fartapp walk refine testdata/walk/ordinary-low-pressure.json --relative-tolerance 1e-8 --max-evaluations 100000
`

func runRefine(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	output, accuracy, err := refinementOptions(args)
	if err != nil {
		writeDiagnostic(stderr, "invalid walk refine: %v\n", err)
		return 1
	}
	if output.help {
		return writeText(stdout, stderr, refineHelp)
	}
	report := readWalkWith(output.positional[0], stdin, func(data []byte) walkcase.Report {
		return walkcase.Refine(data, accuracy)
	})
	report.Operation, report.ImplementationRevision = "refine", walkcase.RefinementRevision
	return writeWalkReport(report, "refine", output.format, stdout, stderr)
}

func refinementOptions(args []string) (outputOptions, coupledblowdown.AccuracyOptions, error) {
	options, values, err := parseValuedOptions(args, "--relative-tolerance", "--max-evaluations", "--absolute-time-tolerance")
	accuracy := coupledblowdown.AccuracyOptions{}
	if err != nil {
		return options, accuracy, err
	}
	if len(options.positional) > 1 || (!options.help && len(options.positional) != 1) {
		return options, accuracy, errors.New("exactly one input path or - is required")
	}
	relative, hasRelative := values["--relative-tolerance"]
	budget, hasBudget := values["--max-evaluations"]
	if !options.help && (!hasRelative || !hasBudget) {
		return options, accuracy, errors.New("--relative-tolerance and --max-evaluations are required")
	}
	if hasRelative {
		accuracy.RelativeTolerance, err = strconv.ParseFloat(relative, 64)
		if err != nil || math.IsNaN(accuracy.RelativeTolerance) || accuracy.RelativeTolerance < 1e-12 || accuracy.RelativeTolerance > 0.1 {
			return options, accuracy, errors.New("--relative-tolerance must be in [1e-12, 0.1]")
		}
	}
	if hasBudget {
		accuracy.MaxEvaluations, err = strconv.Atoi(budget)
		if err != nil || accuracy.MaxEvaluations < 15 || accuracy.MaxEvaluations > coupledblowdown.MaxAccuracyEvaluations {
			return options, accuracy, errors.New("--max-evaluations must be an integer in [15, 1000000]")
		}
	}
	if absolute, exists := values["--absolute-time-tolerance"]; exists {
		accuracy.AbsoluteTimeToleranceSeconds, err = strconv.ParseFloat(absolute, 64)
		if err != nil || math.IsNaN(accuracy.AbsoluteTimeToleranceSeconds) || math.IsInf(accuracy.AbsoluteTimeToleranceSeconds, 0) || accuracy.AbsoluteTimeToleranceSeconds < 0 {
			return options, accuracy, errors.New("--absolute-time-tolerance must be finite and nonnegative")
		}
	}
	return options, accuracy, nil
}
