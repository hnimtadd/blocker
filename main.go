package main

import (
	"blocker/core"
	"blocker/crypto"
	"blocker/network"
	"bytes"
	"fmt"
	"time"
)

func main() {
	trLocal := network.NewLocalTranposrt("LOCAL")
	trRemoteA := network.NewLocalTranposrt("REMOTE_A")
	trRemoteB := network.NewLocalTranposrt("REMOTE_B")
	trRemoteC := network.NewLocalTranposrt("REMOTE_C")

	go func() {
		time.Sleep(time.Second)
		serverA := makeServer("REMOTE_A", trRemoteA, []network.Transport{trLocal}, nil)
		serverA.Start()
	}()

	go func() {
		time.Sleep(time.Second * 2)
		serverB := makeServer("REMOTE_B", trRemoteB, []network.Transport{trRemoteA}, nil)
		serverB.Start()
	}()
	go func() {
		time.Sleep(time.Second * 3)
		serverC := makeServer("REMOTE_C", trRemoteC, []network.Transport{trRemoteB}, nil)
		serverC.Start()
	}()

	go func() {
		time.Sleep(time.Second * 10)
		trLate := network.NewLocalTranposrt("REMOTE_LATE")
		sLate := makeServer(string(trLate.Addr()), trLate, []network.Transport{trRemoteC}, nil)
		sLate.Start()
	}()

	go func() {
		for {
			if err := sendTransaction(trLocal, trRemoteA); err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 2)
		}
	}()

	privKey := crypto.GeneratePrivateKey()
	server := makeServer("LOCAL", trLocal, []network.Transport{}, privKey)
	server.Start()
}

func makeServer(id string, tr network.Transport, seed []network.Transport, privKey *crypto.PrivateKey) *network.Server {
	opt := network.ServerOptions{
		ID:        id,
		PrivKey:   privKey,
		Transport: tr,
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
