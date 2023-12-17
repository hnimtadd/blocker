package core

import (
	"blocker/types"
	"fmt"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestBlockChain(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestNotSignBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	// Random block without signature
	block := RandomBlock(t, uint32(1), types.RandomHash())
	assert.NotNil(t, bc.AddBlock(block))
}

func TestBlockWithNotSignTx(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	// Random block without signature
	block := RandomBlock(t, uint32(1), getPrevBlockHash(t, bc, uint32(0)))
	block.AddTransaction(RandomTx(t))
	assert.NotNil(t, bc.AddBlock(block))
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
	// Random block without signature
	block := RandomBlock(t, uint32(5), types.Hash{})
	assert.NotNil(t, block)
	assert.NotNil(t, bc.AddBlock(block))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))

	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i)))
		assert.Nil(t, bc.AddBlock(block))
		assert.Equal(t, bc.Height(), uint32(i+1))
	}
	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
	assert.NotNil(t, bc.AddBlock(RandomBlock(t, uint32(89), types.Hash{})))
	assert.Nil(t, bc.AddBlock(RandomBlock(t, uint32(1001), getPrevBlockHash(t, bc, uint32(1000)))))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i)))
		assert.Nil(t, bc.AddBlock(block))
		assert.Equal(t, bc.Height(), uint32(i+1))

		header, err := bc.GetHeader(uint32(i + 1))
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestInvalidHeader(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i)))
		assert.Nil(t, bc.AddBlock(block))
		assert.Equal(t, bc.Height(), uint32(i+1))

		header, err := bc.GetHeader(uint32(i + 2))
		assert.NotNil(t, err)
		assert.Nil(t, header)
	}
}

func getPrevBlockHash(t *testing.T, bc *BlockChain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height)
	assert.Nil(t, err)
	assert.NotNil(t, prevHeader)
	return BlockHasher{}.Hash(prevHeader)
}

func TestAddBlockWithInvalidPrevHash(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	lenBlocks := 1000
	for i := 0; i < lenBlocks; i++ {
		block := RandomBlock(t, uint32(i+1), types.Hash{})
		assert.NotNil(t, bc.AddBlock(block))
	}
}

func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	fmt.Println("===>", time.Now().Unix())
	block := RandomBlock(t, 0, types.RandomHash())
	fmt.Println("===>", time.Now().Unix())
	bc, err := NewBlockChain(block, NewInMemoryStorage(), log.NewNopLogger())
	fmt.Println("===>", time.Now().Unix())
	assert.Nil(t, err)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	return bc
}
