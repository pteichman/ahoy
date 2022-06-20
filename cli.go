package ahoy

import (
	"flag"
	"fmt"
	"os"
)

// CLI is the main entry point for the ahoy command line tool.
func CLI(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		return 2
	}

	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}

	return 0
}

type appEnv struct {
	cmd    string
	keygen cmdKeygen
}

func (e *appEnv) fromArgs(args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	switch args[0] {
	case "keygen":
		e.cmd = "keygen"
		return nil

	default:
		return flag.ErrHelp

	}
}

func (e *appEnv) run() error {
	switch e.cmd {
	case "keygen":
		return e.keygen.run()

	default:
		return fmt.Errorf("unknown command: %s", e.cmd)

	}
}

type cmdKeygen struct{}

func (c *cmdKeygen) run() error {
	return nil
}
