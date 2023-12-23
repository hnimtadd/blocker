package wallet

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/types"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var NodeEndpoint = "http://localhost:8080/api/tx"

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

// SendTransactionToNode will send transaction to endpoint with POST http request, transaction should be signed before send over network
func (w *Wallet) SendTransactionToNode(endpoint string, tx *core.Transaction) error {
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", endpoint, buf)
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
		w.transactions = append(w.transactions, tx)
		return nil
	}
	buf.Reset()
	if _, err := io.Copy(buf, res.Body); err != nil {
		return err
	}
	return errors.New(buf.String())
}

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
		TxInner:   transferTx,
		Data:      nil,
		Fee:       fee,
		Nonce:     w.nonce,
		ValidFrom: time.Now().Add(time.Second * 10).UnixNano(),
	}
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	fmt.Printf("send tx with nonce: %d, at: %v\n", tx.Nonce, tx.ValidFrom)
	return w.SendTransactionToNode(NodeEndpoint, tx)
}

func (w *Wallet) NFTMintTransaction(nftType core.NFTAssetType, data []byte, collectionHash types.Hash, metadata map[string]any, fee uint64) error {
	asset := core.NFTAsset{
		Type:       nftType,
		Data:       data,
		Collection: collectionHash,
	}
	metaDataBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(metaDataBuf).Encode(metadata); err != nil {
		return err
	}

	mintTx := core.MintTx{
		NFT:      asset,
		Metadata: metaDataBuf.Bytes(),
	}
	if err := mintTx.Sign(w.privKey); err != nil {
		return err
	}
	tx := &core.Transaction{
		TxInner: mintTx,
		Data:    nil,
		Nonce:   w.nonce,
		Fee:     fee,
	}
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	return w.SendTransactionToNode(NodeEndpoint, tx)
}

func (w *Wallet) CollectionMintTransaction(collectionType core.NFTCollectionType, metadata map[string]any, fee uint64) error {
	collection := core.NFTCollection{
		Type: collectionType,
	}
	metaDataBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(metaDataBuf).Encode(metadata); err != nil {
		return err
	}

	mintTx := core.MintTx{
		NFT:      collection,
		Metadata: metaDataBuf.Bytes(),
	}
	if err := mintTx.Sign(w.privKey); err != nil {
		return err
	}
	tx := &core.Transaction{
		TxInner: mintTx,
		Data:    nil,
		Nonce:   w.nonce,
		Fee:     fee,
	}
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	return w.SendTransactionToNode(NodeEndpoint, tx)
}

func (w *Wallet) DataTransaction(data []byte, fee uint64) error {
	tx := &core.Transaction{
		TxInner: nil,
		Data:    data,
		Nonce:   w.nonce,
		Fee:     fee,
	}
	if err := tx.Sign(w.privKey); err != nil {
		return err
	}
	return w.SendTransactionToNode(NodeEndpoint, tx)
}

func (w *Wallet) GetUserTransaction() []*core.Transaction {
	return w.transactions
}

type UserState struct {
	Addr    string `json:"addr"`
	Balance uint64 `json:"balance"`
	Nonce   uint64 `json:"nonce"`
}

func (w *Wallet) GetUserState() (UserState, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/account/state/%s", w.addr.String()), nil)
	if err != nil {
		return UserState{}, err
	}
	client := http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return UserState{}, err
	}
	userState := new(UserState)
	if err := json.NewDecoder(rsp.Body).Decode(userState); err != nil {
		return UserState{}, err
	}
	return *userState, nil
}
