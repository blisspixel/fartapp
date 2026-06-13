package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	intensity := 1
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil {
			intensity = n
		}
	}
	fmt.Printf("%s (%s)\n", Pick(intensity), Rate(intensity))
}
