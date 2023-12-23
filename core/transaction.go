package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/binary"
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
	Fee       uint64
	Nonce     uint64
}

func (t Transaction) String() string {
	from := "unknown"
	if t.From != nil {
		from = t.From.Address().String()
	}
	return fmt.Sprintf("%s=>[from=%s, Nonce=%v, fee=%v, timestamp=%v]\n", t.Hash(TxHasher{}).Short(), from, t.Nonce, t.Fee, t.timeStamp)
}

// NewNativeTransaction is deprecated, transaction should be created from account
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

// Bytes return data bytes in the transaction
func (tx *Transaction) Bytes() []byte {
	buf := new(bytes.Buffer)

	if tx.Data != nil {
		if err := binary.Write(buf, binary.LittleEndian, tx.Data); err != nil {
			panic(err)
		}
	}

	if tx.TxInner != nil {
		switch txInner := tx.TxInner.(type) {
		case TransferTx:
			if err := binary.Write(buf, binary.LittleEndian, txInner.Bytes()); err != nil {
				panic(err)
			}
		case MintTx:
			if err := binary.Write(buf, binary.LittleEndian, txInner.Bytes()); err != nil {
				panic(err)
			}
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, tx.Nonce); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (tx *Transaction) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(tx.Bytes())
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
		return ErrSigNotExisted
	}
	if !tx.Signature.Verify(tx.From, tx.Bytes()) {
		return ErrSigInvalid
	}
	if tx.TxInner != nil {
		switch ttx := tx.TxInner.(type) {
		case MintTx:
			return ttx.Verify()
		case TransferTx:
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

func (tx *Transaction) ReHash(hasher Hasher[*Transaction]) types.Hash {
	tx.hash = types.Hash{}
	return tx.Hash(hasher)
}

func (tx *Transaction) Copy() *Transaction {
	newTx := &Transaction{
		From:      tx.From,
		Signature: tx.Signature,
		Nonce:     tx.Nonce,
		TxInner:   tx.TxInner,
		timeStamp: tx.timeStamp,
		Data:      tx.Data[:],
		Fee:       tx.Fee,
	}
	return newTx
}
