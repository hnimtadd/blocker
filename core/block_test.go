package core

import (
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHeaderEncodeDecode(t *testing.T) {
	b := RandomBlock(t, 0, types.RandomHash())
	fmt.Println(b.Hash(BlockHasher{}))
}

func TestSignBlock(t *testing.T) {
	b := RandomBlock(t, 0, types.RandomHash())
	var (
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.Public()
	)

	b.Validator = pubKey
	assert.NotNil(t, b.Verify())
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
	assert.Nil(t, b.Verify())
}

func TestSignBlockInvalid(t *testing.T) {
	b := RandomBlock(t, 0, types.RandomHash())
	var (
		privKey        = crypto.GeneratePrivateKey()
		pubKey         = privKey.Public()
		invalidPrivKey = crypto.GeneratePrivateKey()
		invalidPubKey  = invalidPrivKey.Public()
	)
	b.Validator = pubKey
	assert.NotNil(t, b.Verify())
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
	assert.NotNil(t, b.Validator)
	assert.Nil(t, b.Verify())

	b.Validator = invalidPubKey

	assert.NotNil(t, b.Verify())
}

func TestEncodeDecodeBlock(t *testing.T) {
	block := RandomBlock(t, 0, types.RandomHash())
	tx := RandomTxWithSignature(t)
	block.AddTransaction(tx)
	buf := &bytes.Buffer{}
	enc := GobBlockEncoder{
		w: buf,
	}
	dnc := GobBlockDecoder{
		r: buf,
	}
	assert.Nil(t, block.Encode(&enc))
	blockDecoded := &Block{}
	assert.Nil(t, blockDecoded.Decode(&dnc))
	assert.Equal(t, block.Header, blockDecoded.Header)
	assert.Equal(t, block.Height, blockDecoded.Height)
	assert.Equal(t, block.PrevBlockHash, blockDecoded.PrevBlockHash)
	assert.Equal(t, block.DataHash, blockDecoded.DataHash)
	assert.Equal(t, block.Validator, blockDecoded.Validator)
	assert.Equal(t, block.Signature, blockDecoded.Signature)
	assert.Equal(t, len(block.Transactions), len(blockDecoded.Transactions))
	for i := 0; i < len(block.Transactions); i++ {
		assert.Equal(t, block.Transactions[i], blockDecoded.Transactions[i])
	}
}

func RandomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	h := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}
	b, err := NewBlock(h, []*Transaction{})
	assert.Nil(t, err)

	dataHash, err := CalculateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))
	return b
}
