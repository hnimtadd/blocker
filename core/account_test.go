package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodingAccount(t *testing.T) {
	account := NewAccount()

	filePath := "./test/account.txt"
	assert.Nil(t, account.SaveAccount(filePath))
	// assert.Nil(t, account.SaveAccount(filePath))

	loadedAccount, err := LoadAccount(filePath)
	assert.Nil(t, err)
	assert.Equal(t, *account, *loadedAccount)
}
