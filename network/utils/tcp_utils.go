package networkutils

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type TCPType uint8

const (
	TypeStream  TCPType = 0x00
	TypeMessage TCPType = 0x01
)

func Send(to net.Conn, payload []byte) error {
	// Send TypeMessage -> Send Length -> SendMessage
	n := len(payload)
	if err := binary.Write(to, binary.LittleEndian, TypeMessage); err != nil {
		return err
	}
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint16(bytes, uint16(n))
	if err := binary.Write(to, binary.LittleEndian, bytes); err != nil {
		return err
	}

	if err := binary.Write(to, binary.LittleEndian, payload); err != nil {
		return err
	}
	return nil
}

func stream(to net.Conn, r io.Reader) error {
	// Send TypeStream  -> copy from r to conn
	// Send TypeMessage -> Send Length -> SendMessage
	if err := binary.Write(to, binary.LittleEndian, TypeStream); err != nil {
		return err
	}
loop:
	for {
		n, err := io.CopyN(to, r, 512)
		if err != nil {
			if err == io.EOF {
				break loop
			}
			return err
		}
		if n != 512 {
			return fmt.Errorf("given len %d, written %d", 512, n)
		}
	}
	return nil
}

func Receive(from net.Conn) ([]byte, error) {
	t := new(TCPType)
	if err := binary.Read(from, binary.LittleEndian, t); err != nil {
		return nil, err
	}
	switch tcptype := *t; tcptype {
	case TypeMessage:
		return handleMessage(from)
	case TypeStream:
		return handleStream(from)
	default:
		return nil, fmt.Errorf("unsupported type %v", tcptype)
	}
}

func handleMessage(from net.Conn) ([]byte, error) {
	bytes := make([]byte, 8)
	if err := binary.Read(from, binary.LittleEndian, bytes); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint16(bytes)
	payload := make([]byte, length)

	if err := binary.Read(from, binary.LittleEndian, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func handleStream(from net.Conn) ([]byte, error) {
	return nil, nil
}
