package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	pubkeyRe := regexp.MustCompile(`83e(0[1-9]|1[0-2])23$`)

	start := time.Now()
	count := uint64(0)

	valid := func(pubkey []byte) bool {
		atomic.AddUint64(&count, 1)

		// Fail fast if the key doesn't match '83e', so the more expensive
		// check on the hex string happens on fewer candidates.
		suffix := pubkey[len(pubkey)-4:]
		if binary.BigEndian.Uint16(suffix)&0xfff != 0x83e {
			return false
		}

		pubhex := hex.EncodeToString(pubkey)
		return pubkeyRe.MatchString(pubhex)
	}

	keys := make(chan []byte)

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			priv, err := generateMatch(rand.Reader, valid)
			if err != nil {
				log.Printf("generateMatch: %s", err)
				return
			}

			keys <- priv
		}()
	}

	key := <-keys
	
	pub := key[len(key)-ed25519.PublicKeySize:]

	filename := fmt.Sprintf("spring-83-keypair-%s-%x.txt",
		time.Now().Format("2006-01-02"), pub[:6])

	content := fmt.Sprintf("%x\n", key)

	ioutil.WriteFile(filename, []byte(content), 0644)

	fmt.Printf("Checked %d candidates in %s\n", count, time.Since(start).Truncate(time.Millisecond))
	fmt.Printf("Wrote: %s\n", filename)
}

func generateMatch(r io.Reader, valid func([]byte) bool) ([]byte, error) {
	for {
		pub, priv, err := ed25519.GenerateKey(r)
		if err != nil {
			return nil, err
		}

		if valid(pub) {
			return priv, nil
		}
	}
}
