package networkutils

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPUtils(t *testing.T) {
	to, from := net.Pipe()
	payload := []byte("Hello world!")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			assert.Nil(t, Send(to, payload))
		}
		wg.Done()
	}()

	go func() {
		for {
			receivedPayload, err := Receive(from)
			assert.Nil(t, err)
			assert.Equal(t, payload, receivedPayload)
		}
	}()
	wg.Wait()
}
