package main

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/network"
	"blocker/wallet"
	"bytes"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	trLocal := network.NewTCPTransport("LOCAL", ":3000")

	go func() {
		txPostTicker := time.NewTicker(time.Second * 3)
		time.Sleep(time.Second * 2)
		for {
			if err := sendMintTransaction(); err != nil {
				panic(err)
			}
			<-txPostTicker.C
			// fmt.Println("send mint")
		}
	}()

	go func() {
		from := crypto.GeneratePrivateKey()
		to := crypto.GeneratePrivateKey()
		txPostTicker := time.NewTicker(time.Second * 6)
		time.Sleep(time.Second * 2)
		if err := sendTransferTransaction(from, to.Public(), 10, txPostTicker); err != nil {
			panic(err)
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	server := makeServer("localhost:8080", trLocal, []network.Peer{}, privKey)
	server.Start()
}

func makeTCPServer(node network.Transport, tcpSeed []string, privKey *crypto.PrivateKey) *network.Server {
	opt := network.ServerOptions{
		Transport: node,
		ID:        string(node.Addr()),
		PrivKey:   privKey,
		TCPSeed:   tcpSeed,
	}
	server, err := network.NewServer(opt)
	if err != nil {
		panic(err)
	}
	return server
}

func makeServer(apiAddr string, node network.Transport, seed []network.Peer, privKey *crypto.PrivateKey) *network.Server {
	opt := network.ServerOptions{
		Transport: node,
		ID:        string(node.Addr()),
		Addr:      apiAddr,
		PrivKey:   privKey,
		LocalSeed: seed,
	}
	server, err := network.NewServer(opt)
	if err != nil {
		panic(err)
	}
	return server
}

func sendLocalTransaction(to network.Transport, from network.Transport) error {
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewNativeTransaction(data)
	privKey := crypto.GeneratePrivateKey()
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := network.NewMesage(network.MessageTypeTx, buf.Bytes())
	err := to.Send(from.Addr(), msg.Bytes())
	return err
}

func sendDataTransaction() error {
	from := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewNativeTransaction(data)
	if err := tx.Sign(from); err != nil {
		return err
	}
	return sendTransaction(tx)
}

func sendMintTransaction() error {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	from := crypto.GeneratePrivateKey()
	randInt := strconv.Itoa(rand.Int())
	// fmt.Println(randInt)
	mintTx := core.MintTx{
		NFT: core.NFTAsset{
			Type: core.NFTAssetTypeImageBase64,
			Data: []byte(randInt),
		},
	}
	if err := mintTx.Sign(from); err != nil {
		return err
	}

	tx := core.NewNativeMintTransacton(mintTx)
	tx.Nonce = 1
	if err := tx.Sign(from); err != nil {
		return err
	}
	return sendTransaction(tx)
}

func sendTransaction(tx *core.Transaction) error {
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
		return nil
	}
	buf.Reset()
	if _, err := io.Copy(buf, res.Body); err != nil {
		return err
	}
	return errors.New(buf.String())
}

func sendTransferTransaction(from *crypto.PrivateKey, to *crypto.PublicKey, amount uint64, ticker *time.Ticker) error {
	w := wallet.NewWallet(from)
	for {
		<-ticker.C
		if err := w.TransferTransaction(to.Address(), amount, 50); err != nil {
			return err
		}
	}
}
