package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/rand"
)

type (
	TxType string
	MintTx struct {
		// Minting new NFTs belong to existed collection
		NFT       any
		Owner     *crypto.PublicKey
		Signature *crypto.Signature
		Metadata  []byte
		Fee       int32
		hash      types.Hash // cached hash
	}
	// TransferTx struct {
	// 	// Transfer cryptor from wallet to wallet
	// 	In        []TxIn
	// 	Out       []TxOut
	// 	From      crypto.PublicKey
	// 	To        crypto.PublicKey
	// 	Signature crypto.Signature // signature of sender
	// 	Fee       int32
	// }
	// TxIn struct {
	// 	TxID  types.Hash // ID of TxOut
	// 	Index int
	// }
	// TxOut struct {
	// 	Value int
	// }
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

func NewMintTransacton(tx MintTx) *Transaction {
	return &Transaction{
		Nonce:   rand.Uint64(),
		TxInner: tx,
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

func (tx *MintTx) Bytes() []byte {
	buf := new(bytes.Buffer)
	switch nft := tx.NFT.(type) {
	case NFTCollection:
		if err := binary.Write(buf, binary.LittleEndian, []byte(nft.Type)); err != nil {
			panic(err)
		}
	case NFTAsset:
		if err := binary.Write(buf, binary.LittleEndian, nft.Data); err != nil {
			panic(err)
		}
		if err := binary.Write(buf, binary.LittleEndian, []byte(nft.Type)); err != nil {
			panic(err)
		}
		if err := binary.Write(buf, binary.LittleEndian, nft.Collection); err != nil {
			panic(err)
		}
	}
	if err := binary.Write(buf, binary.LittleEndian, tx.Metadata); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (tx *MintTx) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(tx.Bytes())
	tx.Owner = privKey.Public()
	tx.Signature = sig
	return nil
}

func (tx *MintTx) Hash(hasher Hasher[*MintTx]) types.Hash {
	if tx.hash.IsZero() {
		tx.hash = hasher.Hash(tx)
	}
	return tx.hash
}

func (tx *MintTx) Verify() error {
	if tx.Signature == nil || tx.Owner == nil {
		return fmt.Errorf("NFT has no signature")
	}
	if !tx.Signature.Verify(tx.Owner, tx.Bytes()) {
		return fmt.Errorf("invalid NFT signature")
	}
	return nil
}

func (tx *MintTx) Encode(enc Encoder[*MintTx]) error {
	return enc.Encode(tx)
}

func (tx *MintTx) Decode(dnc Decoder[*MintTx]) error {
	return dnc.Decode(tx)
}

func (tx MintTx) String() string {
	return fmt.Sprintf("Owner: %s, sign: %s, nft_type: %v\n", tx.Owner.Address().String(), tx.Signature.String(), tx.NFT)
}

func init() {
	gob.Register(MintTx{})
	// gob.Register(TransferTx{})
}
