package ahoy

import (
	"fmt"
	"io"
)

func CLI(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, "ERROR: No command specified\n")
		return 2
	}

	return 0
}
