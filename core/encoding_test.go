package core

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEncodingWithTransaction(t *testing.T) {
	tx := &Transaction{
		ValidFrom: time.Now().Add(time.Second * 6).UnixNano(),
	}
	buf := new(bytes.Buffer)
	err := tx.Encode(NewGobTxEncoder(buf))
	assert.Nil(t, err)

	decodedTx := new(Transaction)
	err = decodedTx.Decode(NewGobTxDecoder(buf))
	assert.Nil(t, err)

	assert.Equal(t, *tx, *decodedTx)
}
