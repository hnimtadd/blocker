package crypto

import (
	"blocker/types"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	privKeyLen = 64
	pubKeyLen  = 32
	seedLen    = 32
	addressLen = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func GeneratePrivateKeyFromString(str string) *PrivateKey {
	seed, err := hex.DecodeString(str)
	if err != nil {
		panic(fmt.Sprintf("Cannot decode string to seed: %v", err))
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func GeneratePrivateKeyWithSeed(seed []byte) *PrivateKey {
	if len(seed) != seedLen {
		panic(fmt.Sprintf("Seed length not valid, must be %d", seedLen))
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		Value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, pubKeyLen)
	copy(b, p.key[32:])

	return &PublicKey{
		Key: b,
	}
}

type PublicKey struct {
	Key ed25519.PublicKey
}

func (p *PublicKey) Bytes() []byte {
	return p.Key
}

func (p *PublicKey) Address() types.Address {
	b := sha256.Sum256(p.Bytes())
	return types.AddressFromBytes(b[len(b)-addressLen:])
}

type Signature struct {
	Value []byte
}

func (s *Signature) Bytes() []byte {
	return s.Value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.Key, msg, s.Value)
}
