package core

import (
	"blocker/types"
	"fmt"
	"strings"
	"sync"
)

type Storage interface {
	PutBlock(*Block) error
	GetBlock(hash types.Hash) (*Block, error)
	HasBlock(hash types.Hash) bool

	PutCollection(*Transaction) error
	GetCollection(hash types.Hash) (*Transaction, error)
	HasCollection(hash types.Hash) bool

	PutNFT(*Transaction) error
	GetNFT(hash types.Hash) (*Transaction, error)
	HasNFT(hash types.Hash) bool

	PutAccount(*AccountState) error
	GetAccount(types.Address) (*AccountState, error)
	UpdateAccountBalance(types.Address, int) error
	IncreaseAccountNonce(types.Address) error
	AccountStateString() string

	GetTransferOfAccount(addr types.Address) (fromTxx []*Transaction, toTxx []*Transaction, err error)

	PutTransfer(*Transaction) error
	GetTransfer(hash types.Hash) (*Transaction, error)

	GetCoinbaseState() *AccountState
	PutCoinbase(*AccountState) error
}

type InMemoryStorage struct {
	blockState      map[types.Hash]*Block
	collectionState map[types.Hash]*Transaction
	nftState        map[types.Hash]*Transaction
	accountState    map[types.Address]*AccountState
	transferState   map[types.Hash]*Transaction
	coinbase        *AccountState
	lock            sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	store := &InMemoryStorage{
		blockState:      make(map[types.Hash]*Block, 10000),
		collectionState: make(map[types.Hash]*Transaction),
		nftState:        make(map[types.Hash]*Transaction),
		accountState:    make(map[types.Address]*AccountState),
		transferState:   make(map[types.Hash]*Transaction),
	}
	var _ Storage = store
	return store
}

func (s *InMemoryStorage) PutBlock(b *Block) error {
	hash := b.Hash(BlockHasher{})
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.blockState[hash]
	if ok {
		return ErrDocExisted
	}
	s.blockState[hash] = b
	return nil
}

func (r *InMemoryStorage) GetBlock(hash types.Hash) (*Block, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	b, ok := r.blockState[hash]
	if !ok {
		return nil, ErrDocNotExisted
	}
	return b, nil
}

func (r *InMemoryStorage) HasBlock(hash types.Hash) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, ok := r.blockState[hash]
	return ok
}

func (r *InMemoryStorage) PutNFT(tx *Transaction) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	hash := tx.Hash(TxHasher{})
	_, ok := r.nftState[hash]
	if ok {
		return ErrDocExisted
	}
	r.nftState[hash] = tx
	return nil
}

func (r *InMemoryStorage) GetNFT(hash types.Hash) (*Transaction, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	tx, ok := r.nftState[hash]
	if !ok {
		return nil, ErrDocNotExisted
	}
	return tx, nil
}

func (r *InMemoryStorage) HasNFT(hash types.Hash) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, ok := r.nftState[hash]
	return ok
}

func (r *InMemoryStorage) PutCollection(tx *Transaction) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	hash := tx.Hash(TxHasher{})
	_, ok := r.collectionState[hash]
	if ok {
		return ErrDocExisted
	}
	r.collectionState[hash] = tx
	return nil
}

func (r *InMemoryStorage) GetCollection(hash types.Hash) (*Transaction, error) {
	r.lock.Lock()
	tx, ok := r.collectionState[hash]
	r.lock.Unlock()
	if !ok {
		return nil, ErrDocNotExisted
	}
	return tx, nil
}

func (r *InMemoryStorage) HasCollection(hash types.Hash) bool {
	r.lock.Lock()
	_, ok := r.collectionState[hash]
	r.lock.Unlock()
	return ok
}

func (r *InMemoryStorage) PutAccount(acc *AccountState) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	r.accountState[acc.Addr] = acc
	return nil
}

func (r *InMemoryStorage) GetAccount(addr types.Address) (*AccountState, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	acc, ok := r.accountState[addr]
	if !ok {
		acc = NewAccountStateFromAddr(addr)
		r.accountState[addr] = acc
	}
	return acc, nil
}

func (r *InMemoryStorage) PutTransfer(tx *Transaction) error {
	r.lock.RLock()
	defer r.lock.RUnlock()
	r.transferState[tx.Hash(TxHasher{})] = tx
	return nil
}

func (r *InMemoryStorage) GetTransfer(hash types.Hash) (*Transaction, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	tx, ok := r.transferState[hash]
	if !ok {
		return nil, ErrDocNotExisted
	}
	return tx, nil
}

func (r *InMemoryStorage) UpdateAccountBalance(addr types.Address, amount int) error {
	if amount == 0 {
		return nil
	}
	acc, err := r.GetAccount(addr)
	if err != nil {
		return err
	}

	if amount > 0 {
		acc.Balance += uint64(amount)
	} else {
		acc.Balance = uint64(int(acc.Balance) + amount)
	}

	r.accountState[addr] = acc
	return nil
}

func (r *InMemoryStorage) IncreaseAccountNonce(addr types.Address) error {
	acc, err := r.GetAccount(addr)
	if err != nil {
		return err
	}
	r.lock.RLock()
	defer r.lock.RUnlock()
	acc.Nonce += 1
	r.accountState[addr] = acc
	return nil
}

func (r *InMemoryStorage) AccountStateString() string {
	r.lock.Lock()
	acc := r.accountState
	coinbase := r.coinbase
	r.lock.Unlock()

	str := &strings.Builder{}
	str.WriteString("=====================ACCOUNT-STATE=====================\n")
	fmt.Fprintf(str, "coinbase=>%s\n", coinbase)
	for addr, state := range acc {
		fmt.Fprintf(str, "%s=>%s\n", addr.String(), state)
	}
	str.WriteString("=====================END-ACCOUNT-STATE=====================")
	return str.String()
}

func (r *InMemoryStorage) GetTransferOfAccount(addr types.Address) ([]*Transaction, []*Transaction, error) {
	fromTxx := []*Transaction{}
	toTxx := []*Transaction{}
	for _, tx := range r.transferState {
		transfer, ok := tx.TxInner.(TransferTx)
		if !ok {
			continue
		}
		if transfer.From == addr {
			fromTxx = append(fromTxx, tx)
		}
		if transfer.To == addr {
			toTxx = append(toTxx, tx)
		}
	}
	return fromTxx, toTxx, nil
}

func (r *InMemoryStorage) GetCoinbaseState() *AccountState {
	return r.coinbase
}

func (r *InMemoryStorage) PutCoinbase(acc *AccountState) error {
	if r.coinbase != nil {
		return ErrDocExisted
	}
	r.coinbase = acc
	return nil
}
