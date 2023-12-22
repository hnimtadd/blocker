package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"math/rand"
)

type TransferTx struct {
	Signature *crypto.Signature
	Signer    *crypto.PublicKey
	From      types.Address
	To        types.Address
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
	tx.Signer = priv.Public()
	return nil
}

func (tx *TransferTx) Verify() error {
	if tx.Signature == nil || tx.Signer == nil {
		return ErrSigNotExisted
	}
	if !tx.Signature.Verify(tx.Signer, tx.Bytes()) {
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
	return transferTx.Signer == nil &&
		transferTx.From == types.Address{} &&
		transferTx.To == types.Address{} &&
		transferTx.Signature == nil &&
		tx.From == nil &&
		tx.Signature == nil &&
		tx.Nonce == 0
}

func init() {
	gob.Register(TransferTx{})
}
