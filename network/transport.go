package network

type NetAddr string

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SetPeerCh(chan Transport)
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Addr() NetAddr
	Send(NetAddr, []byte) error
}
