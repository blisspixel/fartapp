package repoquality

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type fuzzTarget struct {
	Package string
	Name    string
}

var fuzzTargets = []fuzzTarget{
	{"./internal/cli", "FuzzRun"},
	{"./internal/assurance", "FuzzParse"},
	{"./internal/registrationauthoritybinding", "FuzzCompareExactAuthorityBinding"},
	{"./internal/snapshotregistrationbinding", "FuzzComposePositive"},
	{"./internal/authoritymatching", "FuzzFiniteAuthorityMatching"},
	{"./internal/authorityresolution", "FuzzResolveInSnapshot"},
	{"./internal/cataloglookup", "FuzzFiniteSnapshotAndLookup"},
	{"./internal/catalogregistration", "FuzzRegistrationConstructors"},
	{"./internal/evaluation", "FuzzDispositionConstructors"},
	{"./internal/scenarioprobe", "FuzzValidate"},
	{"./internal/idealmixturereservoir", "FuzzWithdrawFraction"},
	{"./internal/reservoirprediction", "FuzzPredict"},
	{"./internal/restrictionflow", "FuzzEvaluate"},
	{"./internal/restrictionprediction", "FuzzPredict"},
	{"./internal/restrictionhistory", "FuzzIntegrate"},
	{"./internal/restrictionhistoryprediction", "FuzzPredict"},
	{"./internal/coupledblowdown", "FuzzSimulate"},
	{"./internal/walkcase", "FuzzRun"},
	{"./internal/walkcase", "FuzzVerifyRetainedWitnessReport"},
	{"./internal/walkevidence", "FuzzDecode"},
	{"./internal/strictjson", "FuzzInspect"},
	{"./internal/strictjson", "FuzzInspectShape"},
}

func runFuzz(args []string, stdout, stderr io.Writer) int {
	duration := 5 * time.Second
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "-h" || argument == "--help":
			_, _ = io.WriteString(stdout, "usage: repoquality fuzz [--time 5s]\n")
			return 0
		case argument == "--time":
			index++
			if index == len(args) {
				writeDiagnostic(stderr, "fuzz: --time requires a duration\n")
				return 1
			}
			parsed, err := parseDurationArgument(args[index])
			if err != nil {
				writeDiagnostic(stderr, "fuzz: %v\n", err)
				return 1
			}
			duration = parsed
		case len(argument) > 7 && argument[:7] == "--time=":
			parsed, err := parseDurationArgument(argument[7:])
			if err != nil {
				writeDiagnostic(stderr, "fuzz: %v\n", err)
				return 1
			}
			duration = parsed
		default:
			writeDiagnostic(stderr, "fuzz: unknown option %s\n", quote(argument))
			return 1
		}
	}
	root, err := FindRoot(".")
	if err != nil {
		writeDiagnostic(stderr, "fuzz: %v\n", err)
		return 1
	}
	if err := RunFuzz(root, duration, stdout, stderr); err != nil {
		writeDiagnostic(stderr, "fuzz: %v\n", err)
		return 1
	}
	return 0
}

func RunFuzz(root string, duration time.Duration, stdout, stderr io.Writer) error {
	if duration <= 0 {
		return fmt.Errorf("fuzz time must be a positive Go duration")
	}
	for _, target := range fuzzTargets {
		fmt.Fprintf(stdout, "fuzz %s %s\n", target.Package, target.Name)
		command := exec.Command(
			"go", "test", "-run=^$", "-fuzz=^"+regexp.QuoteMeta(target.Name)+"$", "-fuzztime="+duration.String(), target.Package,
		)
		command.Dir = root
		command.Stdout = stdout
		command.Stderr = stderr
		command.Env = os.Environ()
		if err := command.Run(); err != nil {
			return fmt.Errorf("%s %s failed: %w", target.Package, target.Name, err)
		}
	}
	return nil
}
