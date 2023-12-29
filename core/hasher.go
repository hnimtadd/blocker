package core

import (
	"blocker/types"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct{}

func (BlockHasher) Hash(h *Header) types.Hash {
	hash := sha256.Sum256(h.Bytes())
	return types.Hash(hash)
}

type TxHasher struct{}

// TODO: should hash of transaction include from and signature?
func (TxHasher) Hash(tx *Transaction) types.Hash {
	return types.Hash(sha256.Sum256(tx.Bytes()))
}

type TxMintHasher struct{}

func (TxMintHasher) Hash(tx *MintTx) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Bytes())
	return types.Hash(sha256.Sum256(buf.Bytes()))
}
