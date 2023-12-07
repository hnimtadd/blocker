package network

import (
	"blocker/core"
	"bytes"
	"encoding/gob"
	"fmt"
)

type RequestBlocksMessage struct {
	// the height indicate lowest block height to fetch
	From uint32

	// the height indicate highest block height to fetch, 0 mean fetch to latest blocks
	To uint32
}

func (msg *RequestBlocksMessage) Bytes() []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type ResponseBlocksMessage struct {
	Blocks []*core.Block
}

func (msg *ResponseBlocksMessage) Bytes() []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type GetStatusMessage struct {
	// The ID of the requester
	ID string
}

func (msg *GetStatusMessage) Bytes() []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

type StatusMessage struct {
	// the id of the server
	ID            string
	Version       uint32
	CurrentHeight uint32
}

func (msg *StatusMessage) Bytes() []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		panic(fmt.Sprintf("status message: cannot encode to bytes, err: %s", err.Error()))
	}
	return buf.Bytes()
}
