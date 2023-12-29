package core

import (
	"blocker/crypto"
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMintTxSign(t *testing.T) {
	mintTx := MintTx{
		Metadata: []byte("Metadata of the mintTx"),
		NFT: NFTAsset{
			Type: NFTAssetTypeImageBase64,
			Data: []byte("first nft"),
		},
	}
	ownerPriv := crypto.GeneratePrivateKey()
	ownerPubKey := ownerPriv.Public()
	assert.Nil(t, mintTx.Sign(ownerPriv))

	assert.NotNil(t, mintTx.Signature)
	assert.NotNil(t, mintTx.Owner)
	assert.Equal(t, mintTx.Owner, ownerPubKey)
	assert.Nil(t, mintTx.Verify())

	invalidPrivKey := crypto.GeneratePrivateKey()
	invalidPubKey := invalidPrivKey.Public()
	mintTx.Owner = invalidPubKey
	assert.NotNil(t, mintTx.Verify())
}

func TestMintTransaction(t *testing.T) {
	mintTx := MintTx{
		Metadata: []byte("Metadata of the mintTx"),
		NFT: NFTAsset{
			Type: NFTAssetTypeImageBase64,
			Data: []byte("first nft"),
		},
	}
	gob.Register(MintTx{})
	buf := new(bytes.Buffer)
	assert.Nil(t, gob.NewEncoder(buf).Encode(mintTx))
	decodedMintTx := new(MintTx)
	assert.Nil(t, gob.NewDecoder(buf).Decode(decodedMintTx))
	assert.Equal(t, mintTx, *decodedMintTx)
}

func TestTransactionWithCollectionTx(t *testing.T) {
	collectionTx := MintTx{
		NFT:      NFTCollection{},
		Metadata: []byte("new collection"),
	}
	priv := crypto.GeneratePrivateKey()
	tx := Transaction{
		TxInner: collectionTx,
	}
	assert.Nil(t, tx.Sign(priv))
	buf := new(bytes.Buffer)
	assert.Nil(t, gob.NewEncoder(buf).Encode(tx))
	decodedTx := new(Transaction)
	assert.Nil(t, gob.NewDecoder(buf).Decode(decodedTx))
	assert.Equal(t, tx, *decodedTx)
}
