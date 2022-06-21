package spring83

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"runtime"
)

func GenerateKey(ctx context.Context, seeds io.Reader) ([]byte, error) {
	return generateKeyValid(ctx, seeds, ValidPubKey)
}

func GenerateKeyParallel(ctx context.Context, seeds io.Reader) ([]byte, error) {
	nGoroutines := runtime.NumCPU()
	keys := make(chan []byte)

	for i := 0; i < nGoroutines; i++ {
		go func() {
			priv, err := generateKeyValid(ctx, seeds, ValidPubKey)
			if err != nil {
				fmt.Printf("Error in generateKeyValid: %s", err)
				return
			}

			select {
			case keys <- priv:
				close(keys)
			default:
			}
		}()
	}

	key := <-keys
	return key, nil
}

var pubkeyRe = regexp.MustCompile(`83e(01|02|03|04|05|06|07|08|09|10|11|12)2[34]$`)

func ValidPubKey(pub []byte) bool {
	// Fail fast if the key doesn't match '83e', so the more expensive
	// check on the hex string happens on fewer candidates.
	suffix := pub[len(pub)-4:]
	if binary.BigEndian.Uint16(suffix)&0xfff != 0x83e {
		return false
	}

	sufhex := hex.EncodeToString(suffix)
	return pubkeyRe.MatchString(sufhex)
}

// generateKeyValid generates candidate keys in a loop until one is valid
// according to a caller-specified validation function.
//
// ctx can be used to add a timeout or cancel the operation.
//
// It is designed to be used in parallel though it is not parallel itself:
// if seeds can be read concurrently, then GenerateValidKey is thread safe.
//
// seeds should be a source of cryptographically secure random bytes, like
// crypto/rand.Reader.
func generateKeyValid(ctx context.Context, seeds io.Reader, valid func([]byte) bool) ([]byte, error) {
	var priv [ed25519.SeedSize]byte
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if _, err := io.ReadFull(seeds, priv[:]); err != nil {
			return nil, err
		}

		key := ed25519.NewKeyFromSeed(priv[:])
		pub := key[len(key)-ed25519.PublicKeySize:]

		if valid(pub) {
			// priv from GenerateKey is the concatenation of the private
			// key and the public key.
			return key, nil
		}
	}
}
