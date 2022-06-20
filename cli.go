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
	"strconv"
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

	req, err := http.NewRequest("GET", "https://"+c.server+"/"+pub, nil)
	if err != nil {
		return err
	}

	req.Header["User-Agent"] = []string{"ahoy/0.1"}
	req.Header["Spring-Version"] = []string{"83"}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-OK response: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

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
	check := ed25519.NewKeyFromSeed(keypair[:ed25519.SeedSize])
	if !bytes.Equal(check, keypair) {
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

	datefmt := "<time datetime=\"2006-01-02T15:04:05Z\"></time>\n"

	content, err := ioutil.ReadAll(&io.LimitedReader{R: reader, N: 2218})
	if err != nil {
		return err
	}

	if len(content) > 2217 {
		return errors.New("supplied content longer than 2217 bytes")
	}

	now := time.Now().UTC()
	body := append(now.AppendFormat(nil, datefmt), content...)

	if len(body) > 2217 {
		return errors.New("content + date longer than 2217 bytes")
	}

	pub := hex.EncodeToString(keypair[len(keypair)-ed25519.PublicKeySize:])
	sig := hex.EncodeToString(ed25519.Sign(check, body))

	req, err := http.NewRequest("PUT", "https://"+c.server+"/"+pub, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header["User-Agent"] = []string{"ahoy/0.1"}
	req.Header["If-Unmodified-Since"] = []string{now.Format(http.TimeFormat)}

	req.Header["Content-Type"] = []string{"text/html;charset=utf-8"}
	req.Header["Content-Length"] = []string{strconv.Itoa(len(body))}

	req.Header["Spring-Version"] = []string{"83"}
	req.Header["Spring-Signature"] = []string{sig}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return fmt.Errorf("non-OK response: %s", resp.Status)
	}

	return nil
}
