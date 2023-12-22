package core

import (
	"blocker/crypto"
	"blocker/types"
	"fmt"
	"strings"
)

type AccountState struct {
	// PubKey  *crypto.PublicKey
	Addr    types.Address
	Nonce   uint64
	Balance uint64
}

func (s AccountState) String() string {
	str := &strings.Builder{}
	fmt.Fprintf(str, "[balance = %d, nonce = %d]", s.Balance, s.Nonce)
	return str.String()
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
		// PubKey:  pubKey,
		Addr:    pubKey.Address(),
		Balance: 0,
		Nonce:   0,
	}
}

func NewAccountStateFromAddr(addr types.Address) *AccountState {
	return &AccountState{
		Addr:    addr,
		Balance: 0,
		Nonce:   0,
	}
}
