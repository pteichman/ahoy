package ahoy

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func cliUpdate(args []string, stdout, stderr io.Writer) int {
	var app updateApp
	if err := app.fromArgs(args, stdout, stderr); err != nil {
		return 2
	}

	if err := app.run(); err != nil {
		fmt.Fprintf(stderr, "Runtime error: %v\n", err)
		return 1
	}

	return 0
}

type updateApp struct {
	db string

	logger *log.Logger
}

func (app *updateApp) fromArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	flags.SetOutput(stderr)

	flags.StringVar(
		&app.db, "db", "ahoy.db", "Database file",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	app.logger = log.New(stderr, "", log.LstdFlags)

	return nil
}

func (app *updateApp) run() error {
	app.logger.Println("Opening", app.db)

	db, err := sql.Open("sqlite3", app.db)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}
