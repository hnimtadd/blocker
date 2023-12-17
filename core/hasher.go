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

func (TxHasher) Hash(tx *Transaction) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Data)
	binary.Write(buf, binary.LittleEndian, tx.Nonce)
	return types.Hash(sha256.Sum256(buf.Bytes()))
}

type TxMintHasher struct{}

func (TxMintHasher) Hash(tx *MintTx) types.Hash {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tx.Bytes())
	return types.Hash(sha256.Sum256(buf.Bytes()))
}
