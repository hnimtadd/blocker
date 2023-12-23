package core

import (
	"blocker/crypto"
	"blocker/types"
	"fmt"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestSendInsuffienceTransfer(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	validator := crypto.GeneratePrivateKey()
	privBob := crypto.GeneratePrivateKey()
	priveAlice := crypto.GeneratePrivateKey()
	// privHacker := crypto.GeneratePrivateKey()
	assert.Nil(t, bc.store.PutAccount(NewAccountState(privBob.Public())))
	assert.Nil(t, bc.store.PutAccount(NewAccountState(priveAlice.Public())))

	fmt.Println(bc.store.AccountStateString())

	prevHeader, err := bc.GetHeader(0)
	assert.Nil(t, err)
	assert.NotNil(t, prevHeader)

	newBlock := RandomBlock(t, 1, BlockHasher{}.Hash(prevHeader))

	amount := uint64(100)
	transferTx := TransferTx{
		From:  privBob.Public().Address(),
		To:    priveAlice.Public().Address(),
		Value: amount,
	}
	assert.Nil(t, transferTx.Sign(privBob))

	tx := NewNativeTransferTransaction(transferTx)
	assert.Nil(t, tx.Sign(privBob))

	newBlock.AddTransaction(tx)
	assert.Nil(t, newBlock.ReHash(BlockHasher{}))
	assert.Nil(t, newBlock.Sign(validator))

	err = bc.AddBlock(newBlock)
	assert.NotNil(t, err)
	assert.Equal(t, ErrTxInsufficientBalance, err)
	fmt.Println(bc.store.AccountStateString())
}

func TestSendSuccessTransfer(t *testing.T) {
	bc := newBlockChainWithGenesis(t)
	validator := crypto.GeneratePrivateKey()
	privBob := crypto.GeneratePrivateKey()
	priveAlice := crypto.GeneratePrivateKey()
	// privHacker := crypto.GeneratePrivateKey()
	BobState := NewAccountState(privBob.Public())
	BobState.Balance = 1000

	assert.Nil(t, bc.store.PutAccount(BobState))
	assert.Nil(t, bc.store.PutAccount(NewAccountState(priveAlice.Public())))
	fmt.Println(bc.store.AccountStateString())

	prevHeader, err := bc.GetHeader(0)
	assert.Nil(t, err)
	assert.NotNil(t, prevHeader)

	newBlock := RandomBlock(t, 1, BlockHasher{}.Hash(prevHeader))

	amount := uint64(100)
	transferTx := TransferTx{
		From:  privBob.Public().Address(),
		To:    priveAlice.Public().Address(),
		Value: amount,
	}
	assert.Nil(t, transferTx.Sign(privBob))

	tx := NewNativeTransferTransaction(transferTx)
	tx.Fee = 200
	assert.Nil(t, tx.Sign(privBob))

	newBlock.AddTransaction(tx)
	assert.Nil(t, newBlock.ReHash(BlockHasher{}))
	assert.Nil(t, newBlock.Sign(validator))

	assert.Nil(t, bc.AddBlock(newBlock))

	fmt.Println(bc.store.AccountStateString())
}

func TestBlockChain(t *testing.T) {
	newBlockChainWithGenesis(t)
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

func newGenesisBlock() *Block {
	// coinbase := core.Account{}
	transferTx := TransferTx{
		From:  types.Address{},
		To:    types.Address{},
		Value: 1000000,
	}
	tx := NewNativeTransferTransaction(transferTx)

	block := &Block{
		Header: &Header{
			Version:       1,
			PrevBlockHash: types.Hash{},
			Height:        0,
			Timestamp:     00000000,
		},
		Transactions: []*Transaction{
			tx,
		},
	}
	return block
}

func newBlockChainWithGenesis(t *testing.T) *BlockChain {
	fmt.Println("===>", time.Now().Unix())
	block := newGenesisBlock()
	fmt.Println("===>", time.Now().Unix())
	bc, err := NewBlockChain(block, NewInMemoryStorage(), log.NewNopLogger())
	fmt.Println("===>", time.Now().Unix())
	assert.Nil(t, err)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	return bc
}
