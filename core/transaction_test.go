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
		Fee:      200,
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
		Fee:      200,
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
		Fee:      200,
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

func TestSignTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}

	var (
		privKey = crypto.GeneratePrivateKey()
		pubKey  = privKey.Public()
	)
	tx.From = pubKey

	assert.NotNil(t, tx.Verify())

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
	assert.NotNil(t, tx.From)
	assert.Equal(t, tx.From, pubKey)
	assert.Nil(t, tx.Verify())
}

func TestSignInvalidTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}

	var (
		privKey        = crypto.GeneratePrivateKey()
		pubKey         = privKey.Public()
		invalidPrivKey = crypto.GeneratePrivateKey()
		invalidPubKey  = invalidPrivKey.Public()
	)

	assert.Nil(t, tx.Sign(privKey))
	assert.NotNil(t, tx.Signature)
	assert.NotNil(t, tx.From)
	assert.Equal(t, tx.From, pubKey)
	assert.Nil(t, tx.Verify())

	tx.From = invalidPubKey
	assert.NotNil(t, tx.Verify())
}

func TestEncodeDecodeTransaction(t *testing.T) {
	data := []byte("sample")
	tx := Transaction{
		Data: data,
	}
	buf := &bytes.Buffer{}

	enc := &GobTxEncoder{
		w: buf,
	}
	assert.Nil(t, tx.Encode(enc))

	txDecode := Transaction{}
	dnc := &GobTxDecoder{
		r: buf,
	}
	assert.Nil(t, txDecode.Decode(dnc))

	assert.Equal(t, tx, txDecode)
}

func RandomTx(t *testing.T) *Transaction {
	tx := &Transaction{
		Data: []byte("Foo"),
	}
	return tx
}

func RandomTxWithSignature(t *testing.T) *Transaction {
	tx := &Transaction{
		Data: []byte("Foo"),
	}
	privKkey := crypto.GeneratePrivateKey()
	assert.Nil(t, tx.Sign(privKkey))
	return tx
}
