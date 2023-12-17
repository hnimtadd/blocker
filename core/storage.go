package core

import (
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
}

type InMemoryStorage struct {
	blockState      map[types.Hash]*Block
	collectionState map[types.Hash]*Transaction
	nftState        map[types.Hash]*Transaction
	lock            sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	store := &InMemoryStorage{
		blockState:      make(map[types.Hash]*Block, 10000),
		collectionState: make(map[types.Hash]*Transaction),
		nftState:        make(map[types.Hash]*Transaction),
	}
	return store
}

func (s *InMemoryStorage) PutBlock(b *Block) error {
	hash := b.Hash(BlockHasher{})
	s.lock.Lock()
	_, ok := s.blockState[hash]
	s.lock.Unlock()
	if ok {
		return ErrExisted
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
		return nil, ErrNotExisted
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
		return ErrExisted
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
		return nil, ErrNotExisted
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
		return ErrExisted
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
		return nil, ErrNotExisted
	}
	return tx, nil
}

func (r *InMemoryStorage) HasCollection(hash types.Hash) bool {
	r.lock.Lock()
	_, ok := r.collectionState[hash]
	r.lock.Unlock()
	return ok
}
