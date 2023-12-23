package wallet

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Wallet struct {
	privKey      *crypto.PrivateKey
	transactions []*core.Transaction // user's transaction
	addr         types.Address
	nonce        uint64
} // wallet is a tcp node that just hold user-related information, user could init the wallet and attach key to the wallet

func NewWallet(privKey *crypto.PrivateKey) *Wallet {
	w := &Wallet{
		privKey: privKey,
		addr:    privKey.Public().Address(),
		nonce:   1,
	}
	if err := w.RegisterNewWallet(); err != nil {
		panic(err)
	}
	fmt.Printf("created new wallet at addr: %s\n", w.addr.String())
	return w
}

func (w *Wallet) RegisterNewWallet() error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(w.privKey.Public()); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/api/account/register", buf)
	if err != nil {
		return err
	}
	client := http.Client{}
	_, err = client.Do(req)
	return err
}

// TODO: TransferTransaction send transfer transaction, save the transaction information into the inmemory storage
// nonce should be defined from user
func (w *Wallet) TransferTransaction(to types.Address, amount uint64, fee uint64) error {
	transferTx := core.TransferTx{
		From:  w.addr,
		To:    to,
		Value: amount,
	}
	if err := transferTx.Sign(w.privKey); err != nil {
		return err
	}

	tx := &core.Transaction{
		TxInner: transferTx,
		Data:    nil,
		Fee:     fee,
		Nonce:   w.nonce,
	}
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/api/tx", buf)
	if err != nil {
		panic(err)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		w.nonce += 1
		return nil
	}
	buf.Reset()
	if _, err := io.Copy(buf, res.Body); err != nil {
		w.transactions = append(w.transactions, tx)
		return err
	}
	return errors.New(buf.String())
}

func (w *Wallet) GetUserTransaction() []*core.Transaction {
	return w.transactions
}
