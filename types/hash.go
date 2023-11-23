package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

type Hash [32]uint8

func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}
	return true
}

func (h Hash) Bytes() []byte {
	b := make([]byte, 32)
	for i := 0; i < 32; i++ {
		b[i] = h[i]
	}
	return b
}

func (h Hash) String() string {
	return hex.EncodeToString(h.Bytes())
}

// This function return last 4 characters of the hash
func (h Hash) Short() string {
	str := hex.EncodeToString(h.Bytes())
	return str[len(str)-4:]
}

func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic(fmt.Sprintf("Given with length %d, should be 32", len(b)))
	}

	var value [32]uint8
	for i := 0; i < 32; i++ {
		value[i] = b[i]
	}
	return Hash(value)
}

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	io.ReadFull(rand.Reader, token)
	return token
}

func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}
