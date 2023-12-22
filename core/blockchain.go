package core

import (
	"blocker/crypto"
	"blocker/types"
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/log"
)

var (
	ErrTxNotfound            = errors.New("transaction not found")
	ErrHeightTooHigh         = errors.New("given height is too high")
	ErrTxInvalid             = errors.New("given transaction is invalid")
	ErrTxInsufficientBalance = errors.New("given account balance insufficient")
)

type BlockChain struct {
	contractState *State
	logger        log.Logger
	store         Storage
	validator     Validator
	headers       []*Header
	blocks        []*Block
	mintPool      []*TransferTx
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
		mintPool:      make([]*TransferTx, 1000),
		confirmsLevel: 15,
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.handleGenesisBlock(genesis)
	return bc, err
}

func (bc *BlockChain) SetValidator(v Validator) {
	bc.validator = v
}

func (bc *BlockChain) handleGenesisBlock(genesis *Block) error {
	for _, tx := range genesis.Transactions {
		if tx.IsCoinbase() {
			transferTx := tx.TxInner.(TransferTx)
			err := bc.handleCoinbaseTransaction(transferTx)
			if err != nil {
				return nil
			}
			break
		}
	}
	return bc.addBlockWithoutValidation(genesis)
}

func (bc *BlockChain) AddBlock(b *Block) error {
	if err := bc.validator.Validate(b); err != nil {
		return err
	}
	var fee uint64 = 0
	for _, tx := range b.Transactions {
		// logic of vm put here
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}

		// logic of mintTx put here
		if err := bc.handleNatveTransaction(tx); err != nil {
			return err
		}

		// logic of
		fee += tx.Fee
	}

	// TODO: should give fee to minter
	if err := bc.store.UpdateAccountBalance(b.Validator.Address(), int(fee)); err != nil {
		return err
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
	fmt.Println(bc.store.AccountStateString())
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

func (bc *BlockChain) handleNatveTransaction(tx *Transaction) error {
	switch tx.TxInner.(type) {
	case MintTx:
		if err := bc.handleNativeNFTTransaction(tx); err != nil {
			return err
		}
	case TransferTx:
		if err := bc.handleNativeTransferTransaction(tx); err != nil {
			return err
		}
	default:
		return nil
	}
	if err := bc.store.IncreaseAccountNonce(tx.From.Address()); err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) handleNativeNFTTransaction(tx *Transaction) error {
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
			return ErrDocExisted
		}
	default:
		return errors.New("unknow nft inside")
	}

	return nil
}

func (bc *BlockChain) handleCoinbaseTransaction(tx TransferTx) error {
	coinbaseAccount := &AccountState{
		Balance: tx.Value,
	}
	return bc.store.PutCoinbase(coinbaseAccount)
}

func (bc *BlockChain) handleNativeTransferTransaction(tx *Transaction) error {
	transferTx := tx.TxInner.(TransferTx)

	fromState, err := bc.store.GetAccount(transferTx.From)
	if err != nil {
		return err
	}
	if fromState.Balance < (tx.Fee + transferTx.Value) {
		return ErrTxInsufficientBalance
	}
	if err := bc.store.PutTransfer(tx); err != nil {
		return err
	}

	fromTotal := -(int(transferTx.Value) + int(tx.Fee))
	fmt.Println(fromTotal)
	if err := bc.store.UpdateAccountBalance(fromState.Addr, fromTotal); err != nil {
		return err
	}
	if err := bc.store.UpdateAccountBalance(transferTx.To, int(transferTx.Value)); err != nil {
		return err
	}

	return nil
}

// SoftcheckTransactions check list of transaction and return list of index of transactions that not pass the soft check
func (bc *BlockChain) SoftcheckTransactions(txx []*Transaction) []types.Hash {
	idxx := []types.Hash{}
	for _, tx := range txx {
		if err := bc.checkgeneralTransaction(tx); err != nil {
			bc.logger.Log("soft check", err)
			idxx = append(idxx, tx.Hash(TxHasher{}))
			continue
		}

		if tx.TxInner != nil {
			switch tx.TxInner.(type) {
			case MintTx:
				if err := bc.checkNativeNFTTransaction(tx); err != nil {
					bc.logger.Log("soft check mint", err)
					idxx = append(idxx, tx.Hash(TxHasher{}))
					continue
				}
			case TransferTx:
				if err := bc.checkNativeTransferTransaction(tx); err != nil {
					bc.logger.Log("soft check transfer", err)
					idxx = append(idxx, tx.Hash(TxHasher{}))
					continue
				}
			}
		}
	}
	return idxx
}

func (bc *BlockChain) checkgeneralTransaction(tx *Transaction) error {
	fromState, err := bc.store.GetAccount(tx.From.Address())
	if err != nil {
		bc.logger.Log("tx", err)
		return err
	}
	if fromState.Nonce+1 != tx.Nonce {
		return ErrNonceInvalid
	}
	return nil
}

func (bc *BlockChain) checkNativeNFTTransaction(tx *Transaction) error {
	mintTx := tx.TxInner.(MintTx)
	hash := tx.Hash(TxHasher{})
	switch mintTx.NFT.(type) {
	case NFTAsset:
		// logic for mint tx processing should put here
		if ok := bc.store.HasNFT(hash); ok {
			fmt.Println(tx)
			mint, _ := bc.store.GetNFT(hash)
			fmt.Println(mint)
			panic(".")
			return ErrDocExisted
		}

	case NFTCollection:
		// logic for collection tx processing should put here
		if ok := bc.store.HasCollection(hash); ok {
			return ErrDocExisted
		}
	default:
		return errors.New("unknow nft inside")
	}

	return nil
}

func (bc *BlockChain) checkNativeTransferTransaction(tx *Transaction) error {
	transferTx := tx.TxInner.(TransferTx)
	_, err := bc.store.GetAccount(transferTx.From)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) PutNewAccount(pubKey *crypto.PublicKey) error {
	state := NewAccountState(pubKey)
	state.Balance = 1000000 // just for testing
	return bc.store.PutAccount(state)
}
