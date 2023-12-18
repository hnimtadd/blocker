package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
)

var ErrInvalidOwner = errors.New("invalid owner")

type (
	TxIn struct {
		ScriptSig crypto.ScriptSig
		TxID      types.Hash // ID of Tx
		Index     int
	}
	// TxOut will be finded from transaction
	TxOut struct {
		ScriptPub crypto.ScriptPubKey
		Value     uint32
	}
	TransferTx struct {
		// Transfer cryptor from account to account
		Signer    *crypto.PublicKey
		Signature *crypto.Signature // signature of sender
		In        []TxIn
		Out       []TxOut
		Fee       int32
		From      types.Address
		To        types.Address
		hash      types.Hash
	}
)

func NewNativeTransferTransaction(from *crypto.PrivateKey, to types.Address, fee int32, txIn []TxIn, txOut []TxOut) (*Transaction, error) {
	transfer := TransferTx{
		To:  to,
		In:  txIn,
		Out: txOut,
		Fee: fee,
	}
	if err := transfer.Sign(from); err != nil {
		return nil, err
	}
	transaction := &Transaction{
		From:    from.Public(),
		TxInner: transfer,
		Nonce:   rand.Uint64(),
	}
	return transaction, nil
}

func (tx *TxIn) UseKey(pubKey *crypto.PublicKey) bool {
	return tx.ScriptSig.Use(pubKey)
}

func (tx *TxOut) IsLockedWithKey(pubKey *crypto.PublicKey) bool {
	return tx.ScriptPub.Key.Equal(pubKey)
}

func (tx *TransferTx) Bytes() []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, tx.From); err != nil {
		panic(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, tx.To); err != nil {
		panic(err)
	}
	if err := gob.NewEncoder(buf).Encode(tx.In); err != nil {
		panic(err)
	}
	if err := gob.NewEncoder(buf).Encode(tx.Out); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (tx *TransferTx) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(tx.Bytes())
	tx.Signer = privKey.Public()
	tx.Signature = sig
	return nil
}

func (tx *TransferTx) Hash(hasher Hasher[*TransferTx]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *TransferTx) Verify(prevTxs map[types.Hash]*Transaction) error {
	if tx.Signature == nil || tx.Signer == nil {
		return fmt.Errorf("transfer has no signature")
	}
	if !tx.Signature.Verify(tx.Signer, tx.Bytes()) {
		return fmt.Errorf("invalid transfer signature")
	}

	for _, in := range tx.In {
		prevTx, ok := prevTxs[in.TxID]
		if !ok {
			return ErrTxNotfound
		}

		prevTransferTx, ok := prevTx.TxInner.(TransferTx)
		if !ok {
			return ErrTxInvalid
		}

		if in.Index >= len(prevTransferTx.Out) {
			return ErrTxInvalid
		}
		prevTxO := prevTransferTx.Out[in.Index]
		if !crypto.Eval(in.ScriptSig, prevTxO.ScriptPub, tx.Bytes()) {
			return ErrTxInvalid
		}
	}
	return nil
}

func (tx *TransferTx) Encode(enc Encoder[*TransferTx]) error {
	return enc.Encode(tx)
}

func (tx *TransferTx) Decode(dnc Decoder[*TransferTx]) error {
	return dnc.Decode(tx)
}

// func (tx *TxOut) Bytes() []byte {
// 	buf := new(bytes.Buffer)
// 	pubKey := *tx.ScriptPub
// 	if err := binary.Write(buf, binary.LittleEndian, pubKey.Bytes()); err != nil {
// 		panic(err)
// 	}
// 	bufBytes := binary.LittleEndian.AppendUint32(buf.Bytes(), tx.Value)
// 	return bufBytes
// }

// func (tx *TxOut) Spent(privKey *crypto.PrivateKey) error {
// if !tx.CanSpent(privKey.Public()) {
// 	return ErrInvalidOwner
// }
// sig := privKey.Sign(tx.Bytes())
// tx.Signature = sig
// tx.Signer = privKey.Public()
// return nil
// }

// func (tx *TxOut) CanSpent(pubKey *crypto.PublicKey) bool {
// return !tx.isSpent() && tx.Owner == pubKey.Address()
// }
