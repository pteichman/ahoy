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

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "update":
		return cliUpdate(rest, stdout, stderr)

	case "server":
		return cliServer(rest, stdout, stderr)

	default:
		fmt.Fprintf(stderr, "ERROR: Unknown command: %s\n", cmd)
		return 2
	}
}
