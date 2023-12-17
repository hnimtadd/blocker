package network

import (
	"blocker/core"
	"blocker/types"
	"errors"
	"sync"
)

type TxPool struct {
	pending   *TxSortedMap
	processed *TxSortedMap
	denided   *TxSortedMap
	maxLength int // The maxLength of the total pool of transactions. When the pool is full the oldest transaction will be pruned.
}
type TxPoolStatus string

const (
	TxPooldDenied  TxPoolStatus = "denied"
	TxPooldPending TxPoolStatus = "pending"
	TxPoolReceived TxPoolStatus = "received"
	TxPoolUnknown  TxPoolStatus = "unknown"
)

var ErrMaxLengthExceed error = errors.New("max length exceed")

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		processed: NewTxSortedMap(),
		pending:   NewTxSortedMap(),
		maxLength: maxLength,
	}
}

func (p *TxPool) Add(tx *core.Transaction) error {
	// prune the oldest transaction that is sitting in the all pool
	if p.pending.Count() == p.maxLength {
		return ErrMaxLengthExceed
	}
	if !p.pending.Contains(tx.Hash(core.TxHasher{})) {
		p.pending.Add(tx)
	}
	return nil
}

// Contains check if transactions are currently in pool pending pool
func (p *TxPool) Contains(hash types.Hash) bool {
	return p.pending.Contains(hash)
}

/*
Get get transactions from pool and transaction status

- If transaction are in comfirmed pool, that mean the tx is already in the blockchain

- If transaction are in pending pool, that mean the tx is received by the node

- If transaction are in denided pool, that mean the tx is already denided by the blockchain
*/
func (p *TxPool) Get(hash types.Hash) (TxPoolStatus, *core.Transaction, error) {
	tx := p.pending.Get(hash)
	if tx != nil {
		return TxPoolReceived, tx, nil
	}
	tx = p.denided.Get(hash)
	if tx != nil {
		return TxPooldDenied, tx, nil
	}
	tx = p.processed.Get(hash)
	if tx != nil {
		return TxPooldPending, tx, nil
	}
	return TxPoolUnknown, tx, nil
}

// Pending returns a slice of transactions that are in the pending pool
func (p *TxPool) Pending() []*core.Transaction {
	return p.pending.txx.Data
}

func (p *TxPool) ClearPending() {
	p.pending.Clear()
}

func (p *TxPool) PendingCount() int {
	return p.pending.Count()
}

// Denide remove transactions from the current pending
func (p *TxPool) Denide(idxx []types.Hash) {
	for _, idx := range idxx {
		p.pending.Remove(idx)
	}
}

func (p *TxPool) Processed(idxx []types.Hash) {
	for _, idx := range idxx {
		tx := p.pending.Get(idx)
		if tx != nil {
			p.processed.Add(tx)
		}
	}
}

type TxSortedMap struct {
	lookup map[types.Hash]*core.Transaction
	txx    *types.List[*core.Transaction]
	lock   sync.RWMutex
}

func NewTxSortedMap() *TxSortedMap {
	return &TxSortedMap{
		lookup: make(map[types.Hash]*core.Transaction),
		txx:    types.NewList[*core.Transaction](),
	}
}

func (t *TxSortedMap) First() *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	first := t.txx.Get(0)
	return t.lookup[first.Hash(core.TxHasher{})]
}

func (t *TxSortedMap) Get(h types.Hash) *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.lookup[h]
}

func (t *TxSortedMap) Add(tx *core.Transaction) {
	hash := tx.Hash(core.TxHasher{})

	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.lookup[hash]; !ok {
		t.lookup[hash] = tx
		t.txx.Insert(tx)
	}
}

func (t *TxSortedMap) Remove(h types.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.txx.Remove(t.lookup[h])
	delete(t.lookup, h)
}

func (t *TxSortedMap) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.lookup)
}

func (t *TxSortedMap) Contains(h types.Hash) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	_, ok := t.lookup[h]
	return ok
}

func (t *TxSortedMap) Clear() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.lookup = make(map[types.Hash]*core.Transaction)
	t.txx.Clear()
}
