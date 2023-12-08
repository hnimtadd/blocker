package main

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/network"
	"bytes"
	"net/http"
	"time"
)

func main() {
	trLocal := network.NewTCPTransport("LOCAL", ":3000")
	trRemoteA := network.NewTCPTransport("REMOTE_A", ":3001")

	go func() {
		// Late node
		time.Sleep(time.Second * 10)
		serverA := makeTCPServer(trRemoteA, []string{":3000"}, nil)
		serverA.Start()
	}()

	txPostTicker := time.NewTicker(time.Millisecond * 500)

	go func() {
		for {
			if err := sendTransaction(); err != nil {
				panic(err)
			}
			<-txPostTicker.C
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	server := makeServer(":8080", trLocal, []network.Peer{}, privKey)
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
	tx := core.NewTransaction(data)
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

func sendTransaction() error {
	from := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	if err := tx.Sign(from); err != nil {
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
	_, err = client.Do(req)
	return err
}
