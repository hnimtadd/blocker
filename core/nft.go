package core

import (
	"blocker/types"
	"encoding/gob"
)

const (
	NFTAssetTypeImageBase64 NFTAssetType      = "image-base64"
	NFTAssetTypeImageURL    NFTAssetType      = "image-url"
	NFTCollectionTypeImage  NFTCollectionType = "collection-image"
)

type (
	NFTAssetType string
	NFTAsset     struct {
		// NFT assert that could embed into the MintTx
		Type       NFTAssetType
		Data       []byte
		Collection types.Hash // MintTx of collection
	}
)

func (nft *NFTAsset) Encode(enc Encoder[*NFTAsset]) error {
	return enc.Encode(nft)
}

func (nft *NFTAsset) Decode(dec Decoder[*NFTAsset]) error {
	return dec.Decode(nft)
}

func NewNFT(typ NFTAssetType, data []byte) *NFTAsset {
	return &NFTAsset{
		Data: data,
		Type: typ,
	}
}

type (
	NFTCollectionType string
	NFTCollection     struct {
		Type NFTCollectionType
	}
)

func (nft *NFTCollection) Encode(enc Encoder[*NFTCollection]) error {
	return enc.Encode(nft)
}

func (nft *NFTCollection) Decode(dec Decoder[*NFTCollection]) error {
	return dec.Decode(nft)
}

func NewNFTCollection(typ NFTCollectionType) *NFTCollection {
	return &NFTCollection{
		Type: typ,
	}
}

func init() {
	gob.Register(NFTAsset{})
	gob.Register(NFTCollection{})
}
