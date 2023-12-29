package main

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/network"
	"blocker/types"
	"blocker/wallet"
	"bytes"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	trLocal := network.NewTCPTransport("LOCAL", ":3000")

	go func() {
		txPostTicker := time.NewTicker(time.Second * 3)
		time.Sleep(time.Second * 2)
		w := wallet.NewWallet(crypto.GeneratePrivateKey())
		for {
			if err := sendMintTransaction(w); err != nil {
				panic(err)
			}
			<-txPostTicker.C
		}
	}()

	go func() {
		from := crypto.GeneratePrivateKey()
		w := wallet.NewWallet(from)
		to := crypto.GeneratePrivateKey()
		txPostTicker := time.NewTicker(time.Second * 6)
		time.Sleep(time.Second * 2)
		if err := sendTransferTransaction(w, to.Public(), 10, txPostTicker); err != nil {
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
	w := wallet.NewWallet(from)
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	return w.DataTransaction(data, 0)
}

func sendMintTransaction(from *wallet.Wallet) error {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	randInt := strconv.Itoa(rand.Int())
	return from.NFTMintTransaction(
		core.NFTAssetTypeImageBase64,
		[]byte(randInt),
		types.Hash{},
		map[string]any{"name": "hello"},
		0)
}

func sendTransferTransaction(from *wallet.Wallet, to *crypto.PublicKey, amount uint64, ticker *time.Ticker) error {
	for {
		<-ticker.C
		if err := from.TransferTransaction(to.Address(), amount, 50); err != nil {
			return err
		}
	}
}
