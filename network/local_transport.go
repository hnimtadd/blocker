package network

import (
	"bytes"
	"fmt"
	"sync"
)

type LocalTransport struct {
	peers     map[NetAddr]*LocalTransport
	peerCh    chan Transport
	consumeCh chan RPC
	addr      NetAddr
	lock      sync.RWMutex
}

func NewLocalTranposrt(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[NetAddr]*LocalTransport),
	}
}

func (t *LocalTransport) SetPeerCh(peerCh chan Transport) {
	t.peerCh = peerCh
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Connect(tr Transport) error {
	trans := tr.(*LocalTransport)
	t.lock.RLock()
	defer t.lock.RUnlock()
	t.peers[tr.Addr()] = trans
	go func() {
		t.peerCh <- tr
	}()
	return nil
}

func (t *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.addr == to {
		return nil
	}
	peer, ok := t.peers[to]
	if !ok {
		return fmt.Errorf("%s: could not send message to unknow peer %s", t.addr, to)
	}
	peer.consumeCh <- RPC{
		From:    t.addr,
		Payload: bytes.NewReader(payload),
	}
	return nil
}

func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	for _, peer := range t.peers {
		if err := t.SendMessage(peer.addr, payload); err != nil {
			return err
		}
	}
	return nil
}

func (t *LocalTransport) Accept(from Transport) error {
	go func() {
		t.peerCh <- from
	}()
	return nil
}

func (t *LocalTransport) Send(from NetAddr, payload []byte) error {
	if from == t.Addr() {
		return nil
	}
	t.consumeCh <- RPC{
		From:    from,
		Payload: bytes.NewReader(payload),
	}
	return nil
}
