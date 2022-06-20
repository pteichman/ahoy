package ahoy

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
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
	start := time.Now()
	fmt.Printf("Starting key generation. This can take some time.\n")

	key, err := spring83.GenerateKeyParallel(context.Background(), rand.Reader)
	if err != nil {
		return err
	}

	pub := key[len(key)-ed25519.PublicKeySize:]

	filename := fmt.Sprintf("spring-83-keypair-%s-%x.txt",
		time.Now().Format("2006-01-02"), pub[:6])

	content := fmt.Sprintf("%x\n", key)
	if err := ioutil.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated key in %s\n", time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Pubkey: %x\n", pub)
	fmt.Printf("Wrote: %s\n", filename)

	return nil
}
