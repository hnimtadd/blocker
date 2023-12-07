package network

type NetAddr string

type Transport interface {
	Addr() NetAddr
	ConsumeRPC() <-chan RPC
	ConsumePeer() <-chan Peer
	Dial(NetAddr) error
	Send(to NetAddr, payload []byte) error // send payload to the transport
	Broadcast([]byte) error                // Broadcast payload to all peers of this transport
}

type Peer interface {
	SetRPCCh(chan<- RPC)
	Accept(from NetAddr, payload []byte) error // accept payload from the NetAddr send to this peer connect
	Addr() NetAddr                             // addr of the peer
}
