package core

import (
	"blocker/crypto"
)

type AccountState struct {
	PubKey  *crypto.PublicKey
	Nonce   uint64
	Balance uint64
}

type Account struct {
	PrivKey *crypto.PrivateKey
}

func NewAccount(priv *crypto.PrivateKey) *Account {
	return &Account{
		PrivKey: priv,
	}
}

func NewAccountState(pubKey *crypto.PublicKey) *AccountState {
	return &AccountState{
		PubKey:  pubKey,
		Balance: 0,
		Nonce:   0,
	}
}
