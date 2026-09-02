package reservoirprediction

import (
	"errors"
	"io"

	"github.com/blisspixel/fartapp/internal/idealmixturereservoir"
)

var (
	ErrInputTooLarge = errors.New("reservoir prediction input exceeds the byte limit")
	ErrNilInput      = errors.New("reservoir prediction input reader is nil")
)

func ReadBounded(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, ErrNilInput
	}
	limited := io.LimitReader(reader, MaxInputBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxInputBytes {
		return nil, ErrInputTooLarge
	}
	return data, nil
}

func Predict(data []byte) Report {
	if len(data) > MaxInputBytes {
		return failure("FART-E-INPUT-0002", "input", "/", "input_too_large")
	}
	request, diagnostic := parseRequest(data)
	if diagnostic != nil {
		return failure(diagnostic.Code, diagnostic.Stage, diagnostic.Path, diagnostic.ReasonCode)
	}
	transition, err := idealmixturereservoir.WithdrawFraction(
		request.state,
		request.withdrawal,
		request.closure,
	)
	if err != nil {
		return failure("FART-E-MODEL-0002", "model", "/", classifyModelError(err))
	}
	return buildReport(request, transition)
}

func InputFailure(reason string, consultedInputs ...string) Report {
	report := failure("FART-E-IO-0002", "input", "/", reason)
	report.ValidationEnvironment.ConsultedInputs = append([]string(nil), consultedInputs...)
	return report
}

func failure(code, stage, path, reason string) Report {
	return Report{
		Schema:                 ReportSchema,
		Status:                 "invalid",
		ImplementationRevision: ImplementationRevision,
		ValidationEnvironment: ValidationEnvironment{
			ConsultedInputs: []string{"document_bytes"},
			AmbientInputs:   []string{},
		},
		Diagnostics: []Diagnostic{{
			Code: code, Stage: stage, Path: path, ReasonCode: reason,
		}},
	}
}
