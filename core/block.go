package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Header struct {
	Version       uint32
	PrevBlockHash types.Hash
	DataHash      types.Hash
	Height        uint32
	Timestamp     int64
}

func (h *Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	if err := encoder.Encode(h); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	return buf.Bytes()
}

type Block struct {
	*Header
	Validator    *crypto.PublicKey
	Signature    *crypto.Signature
	Transactions []*Transaction
	hash         types.Hash // Cached version of the header hash
}

func NewGenesisBlock() *Block {
	return &Block{
		Header: &Header{
			Version:       0,
			PrevBlockHash: types.Hash{},
			DataHash:      types.Hash{},
			Height:        0,
			Timestamp:     0000000,
		},
		Transactions: []*Transaction{},
	}
}

func NewBlock(h *Header, txs []*Transaction) (*Block, error) {
	return &Block{
		Header:       h,
		Transactions: txs,
	}, nil
}

func NewBlockFromPrevHeader(prevHeader *Header, txx []*Transaction) (*Block, error) {
	dataHash, err := CalculateDataHash(txx)
	if err != nil {
		return nil, err
	}
	header := &Header{
		Version:       1,
		DataHash:      dataHash,
		PrevBlockHash: BlockHasher{}.Hash(prevHeader),
		Height:        prevHeader.Height + 1,
		Timestamp:     time.Now().UnixNano(),
	}

	return NewBlock(header, txx)
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (b *Block) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(b.Bytes())
	b.Validator = privKey.Public()
	b.Signature = sig
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil || b.Validator == nil {
		return fmt.Errorf("Block are not signed")
	}

	if !b.Signature.Verify(b.Validator, b.Bytes()) {
		return fmt.Errorf("cannot verify with pubkey")
	}
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	hash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if hash != b.DataHash {
		return fmt.Errorf("block (%s) has invalid datahash", b.Hash(BlockHasher{}))
	}
	return nil
}

func (b *Block) Encode(encoder Encoder[*Block]) error {
	return encoder.Encode(b)
}

func (b *Block) Decode(decoder Decoder[*Block]) error {
	return decoder.Decode(b)
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func CalculateDataHash(txx []*Transaction) (types.Hash, error) {
	buf := &bytes.Buffer{}
	for _, tx := range txx {
		if err := tx.Encode(NewGobTxEncoder(buf)); err != nil {
			return types.Hash{}, err
		}
	}
	hash := sha256.Sum256(buf.Bytes())
	return hash, nil
}
