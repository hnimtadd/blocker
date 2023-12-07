package main

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/network"
	"bytes"
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

	privKey := crypto.GeneratePrivateKey()
	server := makeServer(trLocal, []network.Peer{}, privKey)
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

func makeServer(node network.Transport, seed []network.Peer, privKey *crypto.PrivateKey) *network.Server {
	opt := network.ServerOptions{
		Transport: node,
		ID:        string(node.Addr()),
		PrivKey:   privKey,
		LocalSeed: seed,
	}
	server, err := network.NewServer(opt)
	if err != nil {
		panic(err)
	}
	return server
}

func sendTransaction(to network.Transport, from network.Transport) error {
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
