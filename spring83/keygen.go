package spring83

import (
	"context"
	"crypto/ed25519"
	"io"
)

// GenerateKey generates candidate keys in a loop until one is valid.
// ctx can be used to add a timeout or cancel the operation.
//
// It is designed to be used in parallel though it is not parallel itself:
// if seeds can be read concurrently and valid can be called concurrently,
// then GenerateValidKey is thread safe.
//
// seeds should be a source of cryptographically secure random bytes, like
// crypto/rand.Reader.
func GenerateKey(ctx context.Context, seeds io.Reader, valid func([]byte) bool) ([]byte, error) {
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
			return priv[:], nil
		}
	}
}
