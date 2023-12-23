package pool

import (
	"blocker/core"
	"blocker/types"
	"errors"
	"fmt"
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

var (
	ErrMaxLengthExceed = errors.New("max length exceed")
	ErrNotknown        = errors.New("unknown error occurs")
)

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
	if !p.pending.Contains(tx.ReHash(core.TxHasher{})) {
		p.pending.Add(tx)
		if tx.Nonce > 1 {
			fmt.Println("new tx to pool", tx)
		}
		return nil
	}
	return ErrNotknown
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

// GentPendingAndLock returns a slice of transactions that are in the pending pool, and then lock the pending

func (p *TxPool) LockPending() {
	p.pending.Lock()
}

func (p *TxPool) UnlockPending() {
	p.pending.Unlock()
}

// Pending returns a slice of transactions that are in the pending pool, and then lock the pending
func (p *TxPool) Pending() []*core.Transaction {
	return p.pending.txx.Data
}

func (p *TxPool) ClearPending() {
	p.pending.Unlock()
	p.pending.Clear()
	fmt.Println("clear")
}

func (p *TxPool) PendingCount() int {
	return p.pending.Count()
}

// Denide unlock the locked pending, remove transactions from the current pending, and lock pending again, this method must run after pending
func (p *TxPool) Denide(idxx []types.Hash) []*core.Transaction {
	p.pending.Unlock()
	denidedTXX := []*core.Transaction{}
	for _, idx := range idxx {
		denidedTXX = append(denidedTXX, p.pending.Remove(idx))
	}
	p.pending.Lock()
	return denidedTXX
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
	hash := tx.ReHash(core.TxHasher{})

	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.lookup[hash]; !ok {
		t.lookup[hash] = tx
		t.txx.Insert(tx)
	}
}

func (t *TxSortedMap) Remove(h types.Hash) *core.Transaction {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.txx.Remove(t.lookup[h])
	tx := t.lookup[h]
	delete(t.lookup, h)
	return tx
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

func (t *TxSortedMap) Lock() {
	t.lock.Lock()
	fmt.Println("lock pool")
}

func (t *TxSortedMap) Unlock() {
	t.lock.Unlock()
	fmt.Println("unlock pool")
}
