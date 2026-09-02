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

type outputOptions struct {
	positional []string
	format     outputFormat
	help       bool
}

func parseOutputOptions(args []string) (outputOptions, error) {
	result := outputOptions{format: outputText}
	seenFormat := false
	seenHelp := false
	optionsEnded := false
	result.positional = make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if optionsEnded {
			result.positional = append(result.positional, argument)
			continue
		}
		if argument == "--" {
			optionsEnded = true
			continue
		}

		var value string
		switch {
		case argument == "-h" || argument == "--help":
			if seenHelp {
				return outputOptions{}, fmt.Errorf("--help may be specified only once")
			}
			seenHelp = true
			result.help = true
			continue
		case argument == "--format":
			if seenFormat {
				return outputOptions{}, fmt.Errorf("--format may be specified only once")
			}
			seenFormat = true
			index++
			if index == len(args) {
				return outputOptions{}, fmt.Errorf("--format requires text or json")
			}
			value = args[index]
		case strings.HasPrefix(argument, "--format="):
			if seenFormat {
				return outputOptions{}, fmt.Errorf("--format may be specified only once")
			}
			seenFormat = true
			value = strings.TrimPrefix(argument, "--format=")
		case strings.HasPrefix(argument, "-") && argument != "-":
			return outputOptions{}, fmt.Errorf("unknown option %s", quoteInput(argument))
		default:
			result.positional = append(result.positional, argument)
			continue
		}

		switch value {
		case "text":
			result.format = outputText
		case "json":
			result.format = outputJSON
		default:
			return outputOptions{}, fmt.Errorf(
				"unsupported format %s; expected text or json",
				quoteInput(value),
			)
		}
	}
	return result, nil
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
