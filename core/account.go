package core

import (
	"blocker/crypto"
	"blocker/types"
	"io/fs"
	"os"
)

// TODO: add account and integrate to blockchain
type Account struct {
	Private *crypto.PrivateKey
	UTXO    map[types.Hash][]int // map unspent transaction of this account
}

func NewAccount() *Account {
	// fresh account
	return &Account{
		Private: crypto.GeneratePrivateKey(),
		UTXO:    make(map[types.Hash][]int),
	}
}

func (a *Account) Encode(enc Encoder[*Account]) error {
	return enc.Encode(a)
}

func (a *Account) Decode(dec Decoder[*Account]) error {
	return dec.Decode(a)
}

func (a *Account) SaveAccount(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, fs.FileMode(0644))
	if err != nil {
		return err
	}
	defer file.Close()

	return a.Encode(NewGobAccountEncoder(file))
}

func LoadAccount(filePath string) (*Account, error) {
	a := &Account{}
	file, err := os.OpenFile(filePath, os.O_RDONLY, fs.FileMode(0666))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := a.Decode(NewGobAccountDecoder(file)); err != nil {
		return nil, err
	}
	return a, nil
}
