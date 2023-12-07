package network

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"net"
)

type SMessage struct {
	FromID string `json:"id"` // ID of the requester
}

func (t SMessage) Bytes() []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&t); err != nil {
		log.Panicf("handshake: cannot encode TCPSynMessage to bytes, err: %s", err.Error())
	}
	return buf.Bytes()
}

func SMessageFromBytes(payload []byte) SMessage {
	msg := new(SMessage)
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(msg); err != nil {
		log.Panicf("handshake: cannot decode TCPSynMessage from bytes, err: %s", err.Error())
	}
	return *msg
}

type SAMessage struct {
	NodeID string `json:"id"` // ID of the responser
}

func (t SAMessage) Bytes() []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&t); err != nil {
		log.Panicf("handshake: cannot encode TCPAckMessage to bytes, err: %s", err.Error())
	}
	return buf.Bytes()
}

func SAMessageFromBytes(payload []byte) SAMessage {
	msg := new(SAMessage)
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(msg); err != nil {
		log.Panicf("handshake: cannot decode SynACKSynMessage from bytes, err: %s", err.Error())
	}
	return *msg
}

type AMessage struct {
	Ack bool
}

func (m AMessage) Bytes() []byte {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(&m); err != nil {
		log.Panicf("handshake: cannot encode AckMessage to bytes, err: %s", err.Error())
	}
	return buf.Bytes()
}

func AMessageFromBytes(payload []byte) AMessage {
	msg := new(AMessage)
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(msg); err != nil {
		log.Panicf("handshake: cannot decode AMessage from bytes, err: %s", err.Error())
	}
	return *msg
}

func DefaultTPCHandshake(fromNode Transport, conn net.Conn) (Node, error) {
	// This method should be conn after dial to other node from fromNode.
	// This should ask the other node the ID of the node and then create peer from this node with that ID
	syn := SMessage{
		FromID: string(fromNode.Addr()),
	}
	n, err := conn.Write(syn.Bytes())
	if err != nil {
		return Node{}, err
	}
	sBytes := syn.Bytes()
	if n != len(sBytes) {
		log.Panicf("given message with len %d, written %d", len(sBytes), n)
	}

	saBytes := make([]byte, 1024)
	n, err = conn.Read(saBytes)
	if err != nil {
		return Node{}, err
	}
	saMessage := SAMessageFromBytes(saBytes[:n])

	aMessage := AMessage{
		Ack: true,
	}
	aBytes := aMessage.Bytes()
	n, err = conn.Write(aBytes)
	if err != nil {
		return Node{}, err
	}
	if n != len(aBytes) {
		log.Panicf("given message with len %d, written %d", len(aBytes), n)
	}

	return Node{
		ID: saMessage.NodeID,
	}, nil
}

func DefaultHandshakeReply(conn net.Conn, to Transport) (Node, error) {
	// This method should call from the transport after receive net connection from readloops
	sBytes := make([]byte, 1024)
	n, err := conn.Read(sBytes)
	if err != nil {
		return Node{}, err
	}

	sMessage := SMessageFromBytes(sBytes[:n])

	saMessage := SAMessage{
		NodeID: string(to.Addr()),
	}

	saBytes := saMessage.Bytes()
	n, err = conn.Write(saBytes)
	if err != nil {
		return Node{}, err
	}
	if n != len(saBytes) {
		log.Panicf("given message with len %d, written %d", len(saBytes), n)
	}

	aBytes := make([]byte, 1024)
	n, err = conn.Read(aBytes)
	if err != nil {
		return Node{}, err
	}
	aMessage := AMessageFromBytes(aBytes[:n])
	if !aMessage.Ack {
		return Node{}, errors.New("handshake fail")
	}
	return Node{ID: sMessage.FromID}, nil
}
