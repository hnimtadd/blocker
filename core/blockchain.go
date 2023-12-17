package core

import (
	"blocker/types"
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

var (
	ErrTxNotfound    = errors.New("transaction not found")
	ErrHeightTooHigh = errors.New("given height is too high")
)

type BlockChain struct {
	contractState *State
	logger        log.Logger
	store         Storage
	validator     Validator
	headers       []*Header
	blocks        []*Block
	confirmsLevel uint32 // number of comfirminations required to consider tx are confirmed
	lock          sync.RWMutex
}

func NewBlockChain(genesis *Block, store Storage, logger log.Logger) (*BlockChain, error) {
	bc := &BlockChain{
		contractState: NewState(),
		logger:        logger,
		store:         store,
		headers:       []*Header{},
		blocks:        []*Block{},
		confirmsLevel: 15,
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
		// logic of vm put here
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}

		// logic of mintTx put here
		if tx.TxInner != nil {
			if err := bc.handleNatveNFTTransaction(tx); err != nil {
				return err
			}
		}

		// logic of
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
		"hash", fmt.Sprintf("%x", b.Hash(BlockHasher{}).Bytes()[:3]),
		"transactions", len(b.Transactions),
	)
	return bc.store.PutBlock(b)
}

func (bc *BlockChain) HasBlock(height uint32) bool {
	return bc.Height() >= height
}

func (bc *BlockChain) GetBlock(height uint32) (*Block, error) {
	if !bc.HasBlock(height) {
		return nil, fmt.Errorf("given height (%d) too high", height)
	}
	bc.lock.Lock()
	block := bc.blocks[height]
	bc.lock.Unlock()
	return block, nil
}

func (bc *BlockChain) GetHeader(height uint32) (*Header, error) {
	if !bc.HasBlock(height) {
		return nil, fmt.Errorf("given height %v too high", height)
	}
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return bc.headers[height], nil
}

func (bc *BlockChain) Height() uint32 {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *BlockChain) GetTransaction(hash types.Hash) (Status, *Block, *Transaction, error) {
	for i := len(bc.blocks) - 1; i >= 0; i-- {
		b := bc.blocks[i]
		for _, tx := range b.Transactions {
			if tx.Hash(TxHasher{}) == hash {
				return bc.statusOfHeight(b.Height), b, tx, nil
			}
		}
	}
	return "", nil, nil, ErrTxNotfound
}

func (bc *BlockChain) statusOfHeight(h uint32) Status {
	if bc.Height()-h > bc.confirmsLevel {
		return StatusConfirmed
	}
	return StatusPending
}

func (bc *BlockChain) handleNatveNFTTransaction(tx *Transaction) error {
	mintTx := tx.TxInner.(MintTx)
	switch mintTx.NFT.(type) {
	case NFTAsset:
		// logic for mint tx processing should put here
		if err := bc.store.PutNFT(tx); err != nil {
			return err
		}

	case NFTCollection:
		// logic for collection tx processing should put here
		if err := bc.store.PutCollection(tx); err != nil {
			return ErrExisted
		}
	default:
		return errors.New("unknow nft inside")
	}

	return nil
}

func (bc *BlockChain) checkNatveNFTTransaction(tx *Transaction) error {
	mintTx := tx.TxInner.(MintTx)
	hash := tx.Hash(TxHasher{})
	switch mintTx.NFT.(type) {
	case NFTAsset:
		// logic for mint tx processing should put here
		if ok := bc.store.HasNFT(hash); ok {
			return ErrExisted
		}

	case NFTCollection:
		// logic for collection tx processing should put here
		if ok := bc.store.HasCollection(hash); ok {
			return ErrExisted
		}
	default:
		return errors.New("unknow nft inside")
	}

	return nil
}

// SoftcheckTransactions check list of transaction and return list of index of transactions that not pass the soft check
func (bc *BlockChain) SoftcheckTransactions(txx []*Transaction) []types.Hash {
	idxx := []types.Hash{}
	for _, tx := range txx {
		if tx.TxInner != nil {
			switch tx.TxInner.(type) {
			case MintTx:
				if err := bc.checkNatveNFTTransaction(tx); err != nil {
					idxx = append(idxx, tx.Hash(TxHasher{}))
				}
			}
		}
	}
	return idxx
}
