package network

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tr1, rpcCh1, peerCh1 := initTCPTransport(":3000")
	_, rpcCh2, peerCh2 := initTCPTransport(":3001")

	payload := []byte("Hello, world!")
	go func() {
		assert.Nil(t, tr1.Dial(":3001"))
		time.Sleep(10 * time.Second)
		tr1.Broadcast(payload)
	}()

	for {
		select {
		case rpc := <-rpcCh1:
			fmt.Println("tcp1 New rpc")
			assert.Equal(t, string(payload), string(rpc.Payload))
		case peer := <-peerCh1:
			fmt.Println("tcp1 new peer", peer.Addr())
		case rpc := <-rpcCh2:
			fmt.Println(string(rpc.Payload))
			fmt.Println("tcp2 new rpc")
			assert.Equal(t, string(payload), string(rpc.Payload))
			return
		case peer := <-peerCh2:
			fmt.Println("tpc2 new peer", peer.Addr())
		}
	}
}

func initTCPTransport(addr NetAddr) (*TCPTransport, <-chan RPC, <-chan Peer) {
	tcpTransport := NewTCPTransport(addr)
	rpcCh := tcpTransport.ConsumeRPC()
	peerCh := tcpTransport.ConsumePeer()
	return tcpTransport, rpcCh, peerCh
}
