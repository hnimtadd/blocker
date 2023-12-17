package core

import (
	"blocker/crypto"
	"blocker/types"
	"fmt"
	"math/rand"
)

type (
	TxType string
)

const (
	TxTypeMint     TxType = "mint"
	TxTypeTransfer TxType = "transfer"
	TxTypeNative   TxType = "native"
)

type Transaction struct {
	// Maybe a native wrapper for any txInner or any state code that run on vm
	From      *crypto.PublicKey
	Signature *crypto.Signature
	TxInner   any
	Data      []byte
	timeStamp int64      // UNIX Nano, first time this transaction be seen locally
	hash      types.Hash // Cached hash version of transaction
	Nonce     uint64
}

func NewNativeTransaction(data []byte) *Transaction {
	return &Transaction{
		Data:  data,
		Nonce: rand.Uint64(), // TODO: find better way to handle Nonce of user
	}
}

func (tx *Transaction) IsTransferTx() bool {
	_, ok := tx.TxInner.(TransferTx)
	return ok
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
		return fmt.Errorf("trannsaction has no signature")
	}
	if !tx.Signature.Verify(tx.From, tx.Data) {
		return fmt.Errorf("invalid tracsaction signature")
	}
	if tx.TxInner != nil {
		switch ttx := tx.TxInner.(type) {
		case MintTx:
			return ttx.Verify()
		default:
			return nil
		}
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
