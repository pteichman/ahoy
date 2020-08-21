package ahoy

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/julienschmidt/httprouter"
)

func cmdServer(args []string, stdout, stderr io.Writer) int {
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

	publicHost string
	publicURL  string

	dbpath string

	logger *log.Logger
}

func (app *serverApp) fromArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(stderr)

	flags.StringVar(
		&app.addr, "addr", "localhost:7411", "Listen address",
	)

	flags.StringVar(
		&app.publicURL, "publicURL", "https://example.org", "Public URL",
	)

	flags.StringVar(
		&app.dbpath, "db", "ahoy.db", "Database file",
	)

	if err := flags.Parse(args); err != nil {
		return err
	}

	app.logger = log.New(stderr, "", log.LstdFlags)

	publicURL, err := url.Parse(app.publicURL)
	if err != nil {
		return fmt.Errorf("publicURL: %s", err)
	}

	app.publicHost = publicURL.Host

	return nil
}

func (app *serverApp) run() error {
	env := &Env{
		PublicHost: app.publicHost,
		PublicURL:  app.publicURL,
		Logger:     log.New(os.Stderr, "", log.LstdFlags),
	}

	router := httprouter.New()
	router.GET("/.well-known/webfinger", handleWebfinger(env))
	router.GET("/users/:username", handleUsers(env))
	router.GET("/actor/:username", handleActor(env))

	srv := http.Server{
		Handler: router,
		Addr:    "localhost:8090",
	}

	env.Logger.Println("starting")
	return srv.ListenAndServe()
}
