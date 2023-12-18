package core

import (
	"blocker/crypto"
	"blocker/types"
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
	GetAccount(*crypto.PublicKey) (*AccountState, error)
	UpdateAccountBalnace(*crypto.PublicKey, int) error
	IncreaseAccountNonce(*crypto.PublicKey) error
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
	return store
}

func (s *InMemoryStorage) PutBlock(b *Block) error {
	hash := b.Hash(BlockHasher{})
	s.lock.Lock()
	_, ok := s.blockState[hash]
	s.lock.Unlock()
	if ok {
		return ErrDocExisted
	}
	s.lock.RLock()
	s.blockState[hash] = b
	s.lock.RUnlock()
	return nil
}

func (r *InMemoryStorage) GetBlock(hash types.Hash) (*Block, error) {
	r.lock.Lock()
	b, ok := r.blockState[hash]
	r.lock.Unlock()
	if !ok {
		return nil, ErrDocNotExisted
	}
	return b, nil
}

func (r *InMemoryStorage) HasBlock(hash types.Hash) bool {
	r.lock.Lock()
	_, ok := r.blockState[hash]
	r.lock.Unlock()
	return ok
}

func (r *InMemoryStorage) PutNFT(tx *Transaction) error {
	r.lock.Lock()
	hash := tx.Hash(TxHasher{})
	_, ok := r.nftState[hash]
	r.lock.Unlock()
	if ok {
		return ErrDocExisted
	}
	r.lock.RLock()
	r.nftState[hash] = tx
	defer r.lock.RUnlock()
	return nil
}

func (r *InMemoryStorage) GetNFT(hash types.Hash) (*Transaction, error) {
	r.lock.Lock()
	tx, ok := r.nftState[hash]
	r.lock.Unlock()
	if !ok {
		return nil, ErrDocNotExisted
	}
	return tx, nil
}

func (r *InMemoryStorage) HasNFT(hash types.Hash) bool {
	r.lock.Lock()
	_, ok := r.nftState[hash]
	r.lock.Unlock()
	return ok
}

func (r *InMemoryStorage) PutCollection(tx *Transaction) error {
	r.lock.Lock()
	hash := tx.Hash(TxHasher{})
	_, ok := r.collectionState[hash]
	r.lock.Unlock()
	if ok {
		return ErrDocExisted
	}
	r.lock.RLock()
	r.collectionState[hash] = tx
	defer r.lock.RUnlock()
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
	r.accountState[acc.PubKey.Address()] = acc
	return nil
}

func (r *InMemoryStorage) GetAccount(pubKey *crypto.PublicKey) (*AccountState, error) {
	addr := pubKey.Address()
	r.lock.Lock()
	defer r.lock.Unlock()
	acc, ok := r.accountState[addr]
	if !ok {
		acc = NewAccountState(pubKey)
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

func (r *InMemoryStorage) UpdateAccountBalnace(pubKey *crypto.PublicKey, amount int) error {
	addr := pubKey.Address()
	r.lock.RLock()
	defer r.lock.RUnlock()
	acc, ok := r.accountState[addr]
	if !ok {
		acc = NewAccountState(pubKey)
	}

	if amount > 0 {
		acc.Balance += uint64(amount)
	} else {
		acc.Balance = uint64(int(acc.Balance) - amount)
	}

	r.accountState[addr] = acc
	return nil
}

func (r *InMemoryStorage) IncreaseAccountNonce(pubKey *crypto.PublicKey) error {
	addr := pubKey.Address()
	r.lock.RLock()
	defer r.lock.RUnlock()
	acc, ok := r.accountState[addr]
	if !ok {
		return ErrDocNotExisted
	}
	acc.Nonce += 1
	r.accountState[addr] = acc
	return nil
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
