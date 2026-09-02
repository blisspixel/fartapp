package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"syscall"
)

type outputFormat uint8

const (
	outputText outputFormat = iota
	outputJSON
)

func parseOutputFormat(args []string) ([]string, outputFormat, error) {
	format := outputText
	seenFormat := false
	optionsEnded := false
	positional := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if optionsEnded {
			positional = append(positional, argument)
			continue
		}
		if argument == "--" {
			optionsEnded = true
			continue
		}

		var value string
		switch {
		case argument == "--format":
			if seenFormat {
				return nil, 0, fmt.Errorf("--format may be specified only once")
			}
			seenFormat = true
			index++
			if index == len(args) {
				return nil, 0, fmt.Errorf("--format requires text or json")
			}
			value = args[index]
		case strings.HasPrefix(argument, "--format="):
			if seenFormat {
				return nil, 0, fmt.Errorf("--format may be specified only once")
			}
			seenFormat = true
			value = strings.TrimPrefix(argument, "--format=")
		case strings.HasPrefix(argument, "-") && argument != "-":
			return nil, 0, fmt.Errorf("unknown option %s", quoteInput(argument))
		default:
			positional = append(positional, argument)
			continue
		}

		switch value {
		case "text":
			format = outputText
		case "json":
			format = outputJSON
		default:
			return nil, 0, fmt.Errorf(
				"unsupported format %s; expected text or json",
				quoteInput(value),
			)
		}
	}
	return positional, format, nil
}

func hasHelpOption(args []string) bool {
	for _, argument := range args {
		if argument == "--" {
			return false
		}
		if argument == "-h" || argument == "--help" {
			return true
		}
	}
	return false
}

func writeText(stdout, stderr io.Writer, output string) int {
	return writeValue(
		stdout,
		stderr,
		outputText,
		output,
		func(value string) string { return value },
	)
}

func writeValue[T any](
	stdout, stderr io.Writer,
	format outputFormat,
	value T,
	textFormatter func(T) string,
) int {
	var output []byte
	if format == outputJSON {
		encoded, err := json.Marshal(value)
		if err != nil {
			writeDiagnostic(stderr, "encode output: %v\n", err)
			return 1
		}
		output = append(encoded, '\n')
	} else {
		output = []byte(textFormatter(value))
	}
	written, err := stdout.Write(output)
	if err == nil && written != len(output) {
		err = io.ErrShortWrite
	}
	if err != nil {
		if !isBrokenPipe(err) {
			writeDiagnostic(stderr, "write output: %v\n", err)
		}
		return 1
	}
	return 0
}

func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) ||
		(runtime.GOOS == "windows" &&
			(errors.Is(err, syscall.Errno(109)) || errors.Is(err, syscall.Errno(232))))
}
