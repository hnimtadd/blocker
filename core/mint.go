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

type MintTx struct {
	// Minting new NFTs belong to existed collection
	NFT       any
	Owner     *crypto.PublicKey
	Signature *crypto.Signature
	Metadata  []byte
	hash      types.Hash // cached hash
}

func NewNativeMintTransacton(tx MintTx) *Transaction {
	return &Transaction{
		Nonce:   rand.Uint64(),
		TxInner: tx,
	}
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
}
