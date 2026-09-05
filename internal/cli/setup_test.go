package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/blisspixel/fartapp/internal/repoquality"
)

type tripwireReader struct{}

func (tripwireReader) Read([]byte) (int, error) {
	panic("help or invalid arguments accessed input")
}

// Fixture arguments are documented relative to the project root. Each package
// has its own test process, so this does not change another package's directory.
func TestMain(m *testing.M) {
	root, err := repoquality.FindRoot(".")
	if err == nil {
		err = os.Chdir(root)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
