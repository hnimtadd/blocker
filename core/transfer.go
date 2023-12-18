package core

import (
	"blocker/crypto"
	"bytes"
	"encoding/binary"
	"math/rand"
)

type TransferTx struct {
	Signature *crypto.Signature
	From      *crypto.PublicKey
	To        *crypto.PublicKey
	Value     uint64
}

func NewNativeTransferTransaction(transferTx TransferTx) *Transaction {
	return &Transaction{
		TxInner: transferTx,
		Nonce:   rand.Uint64(),
	}
}

func (tx *TransferTx) Bytes() []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, tx.Value); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.NativeEndian, tx.To.Bytes()); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (tx *TransferTx) Sign(priv *crypto.PrivateKey) error {
	sig := priv.Sign(tx.Bytes())
	tx.Signature = sig
	tx.From = priv.Public()
	return nil
}

func (tx *TransferTx) Verify() error {
	if tx.Signature == nil || tx.From == nil {
		return ErrSigNotExisted
	}
	if !tx.Signature.Verify(tx.From, tx.Bytes()) {
		return ErrSigInvalid
	}
	return nil
}

func (tx *TransferTx) Encode(enc Encoder[*TransferTx]) error {
	return enc.Encode(tx)
}

func (tx *TransferTx) Decode(dec Decoder[*TransferTx]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) IsCoinbase() bool {
	transferTx, ok := tx.TxInner.(TransferTx)
	if !ok {
		return false
	}
	return transferTx.From == nil &&
		transferTx.To == nil &&
		transferTx.Signature == nil &&
		tx.From == nil &&
		tx.Signature == nil &&
		tx.Nonce == 0
}
