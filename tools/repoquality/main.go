package main

import (
	"io"
	"os"

	"github.com/blisspixel/fartapp/internal/repoquality"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return repoquality.Run(args, stdout, stderr)
}
