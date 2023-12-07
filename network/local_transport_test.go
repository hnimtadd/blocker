package network

// func TestConnect(t *testing.T) {
// 	tra := NewLocalTranposrt("A")
// 	trb := NewLocalTranposrt("B")
// 	assert.Nil(t, tra.Connect(trb))
// 	assert.Nil(t, trb.Connect(tra))
// 	assert.Equal(t, tra.peers[trb.Addr()], trb)
// 	assert.Equal(t, trb.peers[tra.Addr()], tra)
// }

// func TestSendMessage(t *testing.T) {
// 	tra := NewLocalTranposrt("A")
// 	trb := NewLocalTranposrt("B")
// 	assert.Nil(t, tra.Connect(trb))
// 	assert.Nil(t, trb.Connect(tra))
//
// 	msg := []byte("Hello world!")
// 	assert.Nil(t, tra.Send(trb.Addr(), msg))
//
// 	// test send to peer-self
// 	assert.Nil(t, tra.Send(tra.Addr(), msg))
//
// 	// rpc := <-trb.Consume()
// 	// buf := make([]byte, len(msg))
// 	//
// 	// n, err := rpc.Payload.Read(buf)
// 	// assert.Nil(t, err)
// 	// assert.Equal(t, n, len(msg))
// 	//
// 	// assert.Equal(t, msg, buf)
// 	// assert.Equal(t, tra.Addr(), rpc.From)
// 	//
// 	// // Send to invalid peer
// 	// assert.NotNil(t, tra.SendMessage("D", msg))
// }

// func TestBroadcast(t *testing.T) {
// 	tra := NewLocalTranposrt("A")
// 	trb := NewLocalTranposrt("B")
// 	trc := NewLocalTranposrt("C")
// 	assert.Nil(t, tra.Connect(trb))
// 	assert.Nil(t, tra.Connect(trc))
//
// 	msg := []byte("Hello world!")
// 	assert.Nil(t, tra.Broadcast(msg))
//
// 	// buf := make([]byte, len(msg))
// 	// rpcb := <-trb.Consume()
// 	// n, err := rpcb.Payload.Read(buf)
// 	// assert.Nil(t, err)
// 	// assert.Equal(t, n, len(msg))
// 	// assert.Equal(t, msg, buf)
// 	// assert.Equal(t, tra.Addr(), rpcb.From)
// 	//
// 	// buf = make([]byte, len(msg))
// 	// rpcc := <-trc.Consume()
// 	// n, err = rpcc.Payload.Read(buf)
// 	// assert.Nil(t, err)
// 	// assert.Equal(t, n, len(msg))
// 	// assert.Equal(t, msg, buf)
// 	// assert.Equal(t, tra.Addr(), rpcc.From)
// }
