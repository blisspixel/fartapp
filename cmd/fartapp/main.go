package main

import (
	"os"

	"github.com/blisspixel/fartapp/internal/cli"
)

var exitProcess = os.Exit

func main() {
	exitProcess(cli.Run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
