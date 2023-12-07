package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	peerCh chan Peer
	rpcCh  chan RPC
	peers  map[NetAddr]*LocalPeer
	addr   NetAddr
	lock   sync.RWMutex
}

func NewLocalTranposrt(addr NetAddr) *LocalTransport {
	tr := &LocalTransport{
		addr:   addr,
		peers:  make(map[NetAddr]*LocalPeer),
		peerCh: make(chan Peer, 1024),
		rpcCh:  make(chan RPC, 1024),
	}
	var _ Transport = tr
	return tr
}

// func (t *LocalTransport) SetPeerCh(peerCh chan<- Peer) {
// 	t.peerCh = peerCh
// }

func (t *LocalTransport) ConsumeRPC() <-chan RPC {
	return t.rpcCh
}

func (t *LocalTransport) ConsumePeer() <-chan Peer {
	return t.peerCh
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

func (t *LocalTransport) Send(to NetAddr, payload []byte) error {
	if to == t.Addr() {
		return nil
	}
	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("local transport, peer: (%s) not found", to)
	}
	return peer.Accept(t.Addr(), payload)
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		if err := peer.Accept(t.Addr(), payload); err != nil {
			fmt.Printf("error: %s,", fmt.Sprintf("peer send error => addr %s [err: %s]", peer.Addr(), err))
		}
	}
	return nil
}

func (t *LocalTransport) Accept(from NetAddr, payload []byte) error {
	t.rpcCh <- RPC{
		From:    from,
		Payload: payload,
	}
	return nil
}

func (t *LocalTransport) Dial(addr NetAddr) error {
	peer := &LocalPeer{
		rpcCh: t.rpcCh,
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.peers[peer.Addr()] = peer
	t.peerCh <- peer
	return nil
}

type LocalPeer struct {
	rpcCh chan<- RPC
	addr  NetAddr
}

func NewLocalPeer(addr NetAddr) *LocalPeer {
	return &LocalPeer{
		addr: addr,
	}
}

func (p *LocalPeer) SetRPCCh(rpcCh chan<- RPC) {
	p.rpcCh = rpcCh
}

func (p *LocalPeer) Accept(from NetAddr, payload []byte) error {
	p.rpcCh <- RPC{
		From:    from,
		Payload: payload,
	}
	return nil
}

func (p *LocalPeer) Addr() NetAddr {
	return p.addr
}
