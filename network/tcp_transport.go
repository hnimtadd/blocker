package network

import (
	networkutils "blocker/network/utils"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TCPTransport struct {
	listener net.Listener
	peers    map[NetAddr]*TCPPeer
	peerCh   chan Peer
	rpcCh    chan RPC // rpc Chan from the node
	nodeID   string   // Unique identify for this node
	lock     sync.Mutex
}

func NewTCPTransport(nodeID string, addr NetAddr) *TCPTransport {
	listener, err := net.Listen("tcp", string(addr))
	if err != nil {
		panic(fmt.Sprintf("cannot bootstrap the listenr at addr (%s)", string(addr)))
	}
	tcpTransport := &TCPTransport{
		nodeID:   nodeID,
		listener: listener,
		peers:    make(map[NetAddr]*TCPPeer),
		rpcCh:    make(chan RPC, 1024),
		peerCh:   make(chan Peer, 1024),
	}
	var _ Transport = tcpTransport
	go tcpTransport.readLoop()
	return tcpTransport
}

func (t *TCPTransport) ConsumeRPC() <-chan RPC {
	return t.rpcCh
}

func (t *TCPTransport) ConsumePeer() <-chan Peer {
	return t.peerCh
}

func (t *TCPTransport) Send(to NetAddr, payload []byte) error {
	// send the payload to the conn that this tcp are listen to
	// create RPC and send
	t.lock.Lock()
	defer t.lock.Unlock()
	// check if send to ourself
	if to == t.Addr() {
		return nil
	}
	peer, ok := t.peers[to]
	if !ok {
		log.Panicf("[NODE] %s, send to peer [%s] not found, current peers: %v", t.Addr(), to, t.peers)
		return nil
	}
	return peer.Accept(t.Addr(), payload)
}

func (t *TCPTransport) Broadcast(payload []byte) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	// fmt.Printf("[NODE] %s, broadcast msg to %d peer\n", t.Addr(), len(t.peers))
	for _, peer := range t.peers {
		if err := peer.Accept(t.Addr(), payload); err != nil {
			fmt.Printf("[NODE] %s cannot broadcast payload, err: (%s)\n", t.Addr(), err.Error())
		}
	}
	return nil
}

func (t *TCPTransport) Addr() NetAddr {
	return NetAddr(t.nodeID)
}

func (t *TCPTransport) Dial(addr NetAddr) error {
	conn, err := net.Dial("tcp", string(addr))
	if err != nil {
		return err
	}
	node, err := DefaultTPCHandshake(t, conn)
	if err != nil {
		return err
	}
	peer := NewTCPPeer(node.ID, conn, true)
	t.lock.Lock()
	defer t.lock.Unlock()
	t.peers[peer.Addr()] = peer
	fmt.Printf("[NODE] %s add new peer at [%s]\n", t.Addr(), peer.Addr())
	peer.SetRPCCh(t.rpcCh)
	t.peerCh <- peer
	return nil
}

func (t *TCPTransport) readLoop() {
	fmt.Printf("[NODE] %s reading loop\n", t.Addr())
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			panic(fmt.Sprintf("server: cannot accept new conn, err: %s", err.Error()))
		}

		node, err := DefaultHandshakeReply(conn, t)
		if err != nil {
			log.Printf("[NODE] %s, error while reply handshake from conn, err: (%s)\n", t.Addr(), err.Error())
			continue
		}
		peer := NewTCPPeer(node.ID, conn, false)
		fmt.Printf("[NODE] %s add peer from: [%s]\n", t.Addr(), peer.Addr())
		peer.SetRPCCh(t.rpcCh)
		t.peerCh <- peer
		t.peers[peer.Addr()] = peer
	}
}

type TCPPeer struct {
	conn     net.Conn
	rpcCh    chan<- RPC
	nodeID   string // Identify of the node in the other side of the conn
	outbound bool   // outbound peer, if we retrive the conn, then the conn is inbound, else outbound.
}

func NewTCPPeer(nodeID string, conn net.Conn, outbound bool) *TCPPeer {
	tcpPeer := &TCPPeer{
		nodeID:   nodeID,
		conn:     conn,
		outbound: outbound,
	}
	var _ Peer = tcpPeer
	go tcpPeer.readLoop()
	return tcpPeer
}

func (p *TCPPeer) Accept(from NetAddr, payload []byte) error {
	// fmt.Printf("[PEER] %s accept msg from %s\n", p.Addr(), from)
	if from == p.Addr() {
		return nil
	}
	rpc := RPC{
		From:    from,
		Payload: payload,
	}
	rpcBytes := rpc.Bytes()
	err := networkutils.Send(p.conn, rpcBytes)
	return err
	// n, err := io.Copy(p.conn, bytes.NewReader(rpcBytes))
	// if err != nil {
	// 	return err
	// }
	// if int(n) != len(rpcBytes) {
	// 	panic(fmt.Errorf("tcp peer: given payload with len (%d), sent (%d)", len(payload), n))
	// }
	// return nil
}

func (p *TCPPeer) SetRPCCh(rpcCh chan<- RPC) {
	p.rpcCh = rpcCh
}

func (p *TCPPeer) Addr() NetAddr {
	return NetAddr(p.nodeID)
}

func (t *TCPPeer) readLoop() {
	defer t.conn.Close()
	for {
		payload, err := networkutils.Receive(t.conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[PEER] %s, read buf, have io.EOF", t.Addr())
				continue
			}
			panic(fmt.Sprintf("[PEER] %s read from connection failed, err: %s", t.Addr(), err.Error()))
		}
		if len(payload) == 0 {
			panic(fmt.Sprintf("[PEER] %s, empty data from conn", t.Addr()))
		}
		// Buf should be bytes of RPC
		rpc, err := RPCFromBytes(payload)
		if err != nil {
			fmt.Printf("[PEER] %s, error while decoding bytes, err: (%s)", t.Addr(), err.Error())
			panic(".")
			continue
		}
		go func() {
			t.rpcCh <- rpc
		}()
	}
}
