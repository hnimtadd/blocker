package core

import (
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

type BlockChain struct {
	logger    log.Logger
	store     Storage
	headers   []*Header
	blocks    []*Block
	validator Validator
	lock      sync.RWMutex
}

func NewBlockChain(genesis *Block, store Storage, logger log.Logger) (*BlockChain, error) {
	bc := &BlockChain{
		logger:  logger,
		store:   store,
		headers: []*Header{},
		blocks:  []*Block{},
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockWithoutValidation(genesis)
	return bc, err
}

func (bc *BlockChain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *BlockChain) AddBlock(b *Block) error {
	if err := bc.validator.Validate(b); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		bc.logger.Log("msg", "Executing vm", "len", len(tx.Data), "hash", tx.Hash(TxHasher{}))
		vm := NewVM(tx.Data)
		if err := vm.Run(); err != nil {
			return err
		}
		bc.logger.Log("vm result", vm.stack.Pop())
	}

	return bc.addBlockWithoutValidation(b)
}

func (bc *BlockChain) addBlockWithoutValidation(b *Block) error {
	bc.lock.RLock()
	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	bc.lock.RUnlock()

	bc.logger.Log(
		"msg", "new block",
		"height", b.Height,
		"hash", b.Hash(BlockHasher{}),
		"transactions", len(b.Transactions),
	)
	return bc.store.Put(b)
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return height <= bc.Height()
}

func (bc *BlockChain) GetBlock(height uint32) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if !bc.HasBlock(height) {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}
	return bc.blocks[height], nil
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if !bc.HasBlock(height) {
		return nil, fmt.Errorf("given height %v too high", height)
	}
	return bc.headers[height], nil
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}
