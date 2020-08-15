package ahoy

import (
	"flag"
	"fmt"
	"io"
	"log"
)

func cliServer(args []string, stdout, stderr io.Writer) int {
	var app serverApp
	if err := app.fromArgs(args, stdout, stderr); err != nil {
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(stderr, "Runtime error: %v\n", err)
		return 1
	}

	return 0
}

type serverApp struct {
	addr string
	db   string

	logger *log.Logger
}

func (app *serverApp) fromArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(stderr)

	flags.StringVar(
		&app.addr, "addr", "localhost:7411", "Listen address",
	)

	flags.StringVar(
		&app.db, "db", "ahoy.db", "Database file",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	app.logger = log.New(stderr, "", log.LstdFlags)

	return nil
}

func (app *serverApp) run() error {
	return nil
}
