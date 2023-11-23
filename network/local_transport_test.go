package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tra := NewLocalTranposrt("A")
	trb := NewLocalTranposrt("B")
	tra.Connect(trb)
	trb.Connect(tra)
	assert.Equal(t, tra.peers[trb.Addr()], trb)
	assert.Equal(t, trb.peers[tra.Addr()], tra)
}

func TestSendMessage(t *testing.T) {
	tra := NewLocalTranposrt("A")
	trb := NewLocalTranposrt("B")
	tra.Connect(trb)
	trb.Connect(tra)

	msg := []byte("Hello world!")
	assert.Nil(t, tra.SendMessage(trb.Addr(), msg))

	// test send to peer-self
	assert.Nil(t, tra.SendMessage(tra.Addr(), msg))

	rpc := <-trb.Consume()
	buf := make([]byte, len(msg))

	n, err := rpc.Payload.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, n, len(msg))

	assert.Equal(t, msg, buf)
	assert.Equal(t, tra.Addr(), rpc.From)

	// Send to invalid peer
	assert.NotNil(t, tra.SendMessage("D", msg))
}

func TestBroadcast(t *testing.T) {
	tra := NewLocalTranposrt("A")
	trb := NewLocalTranposrt("B")
	trc := NewLocalTranposrt("C")
	tra.Connect(trb)
	tra.Connect(trc)

	msg := []byte("Hello world!")
	assert.Nil(t, tra.Broadcast(msg))

	buf := make([]byte, len(msg))
	rpcb := <-trb.Consume()
	n, err := rpcb.Payload.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, n, len(msg))
	assert.Equal(t, msg, buf)
	assert.Equal(t, tra.Addr(), rpcb.From)

	buf, n, err = make([]byte, len(msg)), 0, nil
	rpcc := <-trc.Consume()
	n, err = rpcc.Payload.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, n, len(msg))
	assert.Equal(t, msg, buf)
	assert.Equal(t, tra.Addr(), rpcc.From)
}
