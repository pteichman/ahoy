package ahoy

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pteichman/ahoy/spring83"
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
	cmd string

	get    cmdGet
	keygen cmdKeygen
	put    cmdPut
}

func (e *appEnv) fromArgs(args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	switch args[0] {
	case "get":
		e.cmd = "get"

		fs := flag.NewFlagSet("get", flag.ContinueOnError)
		fs.StringVar(&e.get.server, "server", "bogbody.biz", "Spring '83 server hostname")
		fs.StringVar(&e.get.keypair, "keypair", "", "Spring '83 keypair filename")

		return fs.Parse(args[1:])

	case "keygen":
		e.cmd = "keygen"
		return nil

	case "put":
		e.cmd = "put"

		fs := flag.NewFlagSet("put", flag.ContinueOnError)
		fs.StringVar(&e.put.server, "server", "bogbody.biz", "Spring '83 server hostname")
		fs.StringVar(&e.put.keypair, "keypair", "", "Spring '83 keypair filename")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		e.put.filename = fs.Arg(0)
		if e.put.filename == "" {
			e.put.filename = "-"
		}

		return nil

	default:
		return flag.ErrHelp

	}
}

func (e *appEnv) run() error {
	switch e.cmd {
	case "get":
		return e.get.run()

	case "keygen":
		return e.keygen.run()

	case "put":
		return e.put.run()

	default:
		return fmt.Errorf("unknown command: %s", e.cmd)

	}
}

type cmdGet struct {
	server  string
	keypair string
}

func (c *cmdGet) run() error {
	keypairtxt, err := ioutil.ReadFile(c.keypair)
	if err != nil {
		return err
	}

	if len(keypairtxt) < ed25519.PrivateKeySize*2 {
		return errors.New("short hex-encoded keypair")
	}

	keypair, err := hex.DecodeString(string(keypairtxt)[:ed25519.PrivateKeySize*2])
	if err != nil {
		return err
	}

	// Recalculate and check the public key to make sure the keypair is legit.
	check := ed25519.NewKeyFromSeed(keypair[:ed25519.SeedSize])
	if !bytes.Equal(check, keypair) {
		return errors.New("invalid keypair")
	}

	pub := hex.EncodeToString(keypair[len(keypair)-ed25519.PublicKeySize:])

	board, err := spring83.Get(*http.DefaultClient, c.server, pub)
	if err != nil {
		return err
	}

	fmt.Println(string(board))

	return nil
}

type cmdKeygen struct{}

func (c *cmdKeygen) run() error {
	start := time.Now()
	fmt.Printf("Starting key generation. This can take some time.\n")

	keypair, err := spring83.GenerateKeyParallel(context.Background(), rand.Reader)
	if err != nil {
		return err
	}

	pub := keypair[len(keypair)-ed25519.PublicKeySize:]

	filename := fmt.Sprintf("spring-83-keypair-%s-%x.txt",
		time.Now().Format("2006-01-02"), pub[:6])

	content := fmt.Sprintf("%x\n", keypair)
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated key in %s\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Pubkey: %x\n", pub)
	fmt.Printf("Wrote: %s\n", filename)

	return nil
}

type cmdPut struct {
	server   string
	keypair  string
	filename string
}

func (c *cmdPut) run() error {
	keypairtxt, err := ioutil.ReadFile(c.keypair)
	if err != nil {
		return err
	}

	if len(keypairtxt) < ed25519.PrivateKeySize*2 {
		return errors.New("short hex-encoded keypair")
	}

	keypair, err := hex.DecodeString(string(keypairtxt)[:ed25519.PrivateKeySize*2])
	if err != nil {
		return err
	}

	// Recalculate and check the public key to make sure the keypair is legit.
	priv := ed25519.NewKeyFromSeed(keypair[:ed25519.SeedSize])
	if !bytes.Equal(priv, keypair) {
		return errors.New("invalid keypair")
	}

	var reader io.Reader
	if c.filename == "-" {
		reader = os.Stdin
	} else {
		reader, err = os.Open(c.filename)
		if err != nil {
			return err
		}
	}

	board, err := ioutil.ReadAll(&io.LimitedReader{R: reader, N: spring83.MaxBoardLen + 1})
	if err != nil {
		return err
	}

	if len(board) > spring83.MaxBoardLen {
		return errors.New("supplied content longer than 2217 bytes")
	}

	now := time.Now().UTC()
	board = append(now.AppendFormat(nil, spring83.BoardDateFormat), board...)

	return spring83.Put(c.server, priv, now, board)
}
