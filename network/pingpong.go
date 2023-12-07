package network

import "net"

const (
	pingMessage = "ping"
	pongMessage = "pong"
)

func Ping(conn net.Conn) bool {
	ping := []byte(pingMessage)
	n, err := conn.Write(ping)
	if err != nil {
		return false
	}
	if n != len(ping) {
		return false
	}
	return true
}

func Pong(conn net.Conn) bool {
	ping := make([]byte, 10)
	n, err := conn.Read(ping)
	if err != nil {
		return false
	}
	if string(ping[:n]) != pingMessage {
		return false
	}
	pong := []byte(pongMessage)
	n, err = conn.Write(pong)
	if err != nil {
		return false
	}
	if n != len(pong) {
		return false
	}
	return true
}
