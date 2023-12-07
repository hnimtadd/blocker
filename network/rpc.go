package network

import (
	"blocker/core"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type MessageType byte

const (
	MessageTypeTx             MessageType = 0x1
	MessageTypeBlock          MessageType = 0x2
	MessageTypeRequestBlocks  MessageType = 0x3
	MessageTypeResponseBlocks MessageType = 0x4
	MessageTypeRequestStatus  MessageType = 0x5
	MessageTypeResponseStatus MessageType = 0x6
)

type RPC struct {
	From    NetAddr
	Payload []byte
}

func (rpc *RPC) Bytes() []byte {
	b := &bytes.Buffer{}

	if err := gob.NewEncoder(b).Encode(rpc); err != nil {
		panic(fmt.Sprintf("rpc: cannot encode to bytes, err: %s", err.Error()))
	}
	return b.Bytes()
}

func RPCFromBytes(buf io.Reader) RPC {
	rpc := &RPC{}
	if err := gob.NewDecoder(buf).Decode(rpc); err != nil {
		panic(fmt.Sprintf("rpc: cannot decode from bytes, err: %s", err.Error()))
	}
	return RPC{
		From:    rpc.From,
		Payload: rpc.Payload,
	}
}

type Message struct {
	Data   []byte
	Header MessageType
}

type DecodedMessage struct {
	Data any
	From NetAddr
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultDecodeMessageFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(bytes.NewBuffer(rpc.Payload)).Decode(&msg); err != nil {
		return nil, fmt.Errorf("cannot decode message from rpc %s: %s", rpc.From, err)
	}

	logrus.WithFields(logrus.Fields{
		"from": rpc.From,
		"type": msg.Header,
	}).Debug("New incomming message")

	switch msg.Header {
	case MessageTypeTx:
		// For transaction
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		// For block
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil

	case MessageTypeRequestBlocks:
		requestMessage := new(RequestBlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(requestMessage); err != nil {
			return nil, fmt.Errorf("cannot decode request blocks message, err: %s", err.Error())
		}
		// For Request Blocks
		return &DecodedMessage{
			From: rpc.From,
			Data: requestMessage,
		}, nil

	case MessageTypeResponseBlocks:
		// For Response Blocks
		responseMessage := new(ResponseBlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(responseMessage); err != nil {
			return nil, fmt.Errorf("cannot decode reponse blocks message, err: %s", err.Error())
		}
		// For Request Blocks
		return &DecodedMessage{
			From: rpc.From,
			Data: responseMessage,
		}, nil

	case MessageTypeRequestStatus:
		requestMessage := new(GetStatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(requestMessage); err != nil {
			return nil, fmt.Errorf("cannot decode request status message, err: %s", err.Error())
		}
		// For Request Blocks
		return &DecodedMessage{
			From: rpc.From,
			Data: requestMessage,
		}, nil
	case MessageTypeResponseStatus:
		responseMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(responseMessage); err != nil {
			return nil, fmt.Errorf("cannot decode response status message, err: %s", err.Error())
		}
		// For Request Blocks
		return &DecodedMessage{
			From: rpc.From,
			Data: responseMessage,
		}, nil

	default:
		return nil, fmt.Errorf("invalid message header: %s", string(msg.Header))
	}
}

func NewMesage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		panic(fmt.Sprintf("Cannot encode messages: %s", err))
	}
	return buf.Bytes()
}

type RPCHandler interface {
	Handle(RPC) error
}

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
