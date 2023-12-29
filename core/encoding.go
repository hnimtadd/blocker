package core

import (
	"encoding/gob"
	"io"
)

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobTxEncoder struct {
	w io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	return &GobTxEncoder{
		w: w,
	}
}

func (enc *GobTxEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(enc.w).Encode(tx)
}

type GobTxDecoder struct {
	r io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	return &GobTxDecoder{
		r: r,
	}
}

func (enc *GobTxDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(enc.r).Decode(tx)
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

func (enc *GobBlockEncoder) Encode(b *Block) error {
	return gob.NewEncoder(enc.w).Encode(b)
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

func (enc *GobBlockDecoder) Decode(b *Block) error {
	return gob.NewDecoder(enc.r).Decode(b)
}

type GobNFTEncoder struct {
	w io.Writer
}

func NewGobNFTEncoder(w io.Writer) *GobNFTEncoder {
	return &GobNFTEncoder{
		w: w,
	}
}

func (enc *GobNFTEncoder) Encode(nft *NFTAsset) error {
	return gob.NewEncoder(enc.w).Encode(nft)
}

type GobNFTDecoder struct {
	r io.Reader
}

func NewGobNFTDecoder(r io.Reader) *GobNFTDecoder {
	return &GobNFTDecoder{
		r: r,
	}
}

func (dec *GobNFTDecoder) Decode(nft *NFTAsset) error {
	return gob.NewDecoder(dec.r).Decode(nft)
}

type GobMintTxEncoder struct {
	w io.Writer
}

func NewGobMintTxEncoder(w io.Writer) *GobMintTxEncoder {
	return &GobMintTxEncoder{
		w: w,
	}
}

func (enc *GobMintTxEncoder) Encode(tx *MintTx) error {
	return gob.NewEncoder(enc.w).Encode(tx)
}

type GobMintTxDecoder struct {
	r io.Reader
}

func NewGobMintTxDecoder(r io.Reader) *GobMintTxDecoder {
	return &GobMintTxDecoder{
		r: r,
	}
}

func (enc *GobMintTxDecoder) Decode(tx *MintTx) error {
	return gob.NewDecoder(enc.r).Decode(tx)
}

type GobTransferTxEncoder struct {
	w io.Writer
}

func NewGobTransferTxEncoder(w io.Writer) *GobTransferTxEncoder {
	return &GobTransferTxEncoder{
		w: w,
	}
}

func (enc *GobTransferTxEncoder) Encode(tx *TransferTx) error {
	return gob.NewEncoder(enc.w).Encode(tx)
}

type GobTransferTxDecoder struct {
	r io.Reader
}

func NewGobTransferTxDecoder(r io.Reader) *GobTransferTxDecoder {
	return &GobTransferTxDecoder{
		r: r,
	}
}

func (enc *GobTransferTxDecoder) Decode(tx *TransferTx) error {
	return gob.NewDecoder(enc.r).Decode(tx)
}
