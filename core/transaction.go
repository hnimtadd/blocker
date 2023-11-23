package core

import (
	"blocker/crypto"
	"blocker/types"
	"fmt"
)

type Transaction struct {
	Data      []byte
	From      *crypto.PublicKey
	Signature *crypto.Signature

	// Cached hash version of transaction
	hash types.Hash
	// UNIX Nano, first time this transaction be seen locally
	timeStamp int64
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}
func (tx *Transaction) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(tx.Data)
	tx.From = privKey.Public()
	tx.Signature = sig
	return nil
}

func (tx *Transaction) Hash(hasher Hasher[*Transaction]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil || tx.From == nil {
		return fmt.Errorf("Trannsaction has no signature")
	}
	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("Invalid tracsaction signature")
	}
	return nil
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}
func (tx *Transaction) Decode(dnc Decoder[*Transaction]) error {
	return dnc.Decode(tx)
}

func (tx *Transaction) SetTimestamp(t int64) {
	tx.timeStamp = t
}

func (tx *Transaction) Timestamp() int64 {
	return tx.timeStamp
}
