package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: fartapp <intensity>")
		return 1
	}
	intensity, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(stderr, "invalid intensity %q: must be an integer from 1 to 5\n", args[1])
		return 1
	}
	if intensity < 1 || intensity > 5 {
		fmt.Fprintf(stderr, "invalid intensity %d: must be from 1 to 5\n", intensity)
		return 1
	}
	fmt.Fprintf(stdout, "%s (%s)\n", Pick(intensity), Rate(intensity))
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}
