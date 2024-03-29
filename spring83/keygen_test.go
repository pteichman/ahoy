package spring83

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"testing"
)

func BenchmarkKeygen_GenerateKey(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			generateKeyValid(context.Background(), rand.Reader, func([]byte) bool { return true })
		}
	})
}

func BenchmarkKeygen_GenerateKeyInsecure(b *testing.B) {
	// A bad actor may want to flood a system with valid keys that are not
	// cryptographically secure. Benchmark how quickly those can be
	// generated.
	b.RunParallel(func(pb *testing.PB) {
		r := &incrReader{cur: 0, incr: 17}
		for pb.Next() {
			generateKeyValid(context.Background(), r, func([]byte) bool { return true })
		}
	})
}

func BenchmarkKeygen_ValidPubKey(b *testing.B) {
	buf := make([]byte, ed25519.PublicKeySize)
	for i := 0; i < b.N; i++ {
		ValidPubKey(buf)
	}
}

type incrReader struct {
	cur  uint64
	incr uint64
	buf  [8]byte
}

func (r *incrReader) Read(p []byte) (int, error) {
	count := len(p)
	for len(p) > 8 {
		binary.BigEndian.PutUint64(r.buf[:], r.cur)
		copy(p, r.buf[:])
		p = p[8:]

		r.cur += r.incr
	}

	binary.BigEndian.PutUint64(r.buf[:], r.cur)
	copy(p, r.buf[:len(p)])
	r.cur += r.incr
	return count, nil
}
