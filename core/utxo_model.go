package core

import (
	"blocker/crypto"
	"blocker/types"
)

func (bc *BlockChain) checkNativeTransferTransaction(tx *Transaction) error {
	transferTx := tx.TxInner.(TransferTx)
	// Check if txxIn are valid
	prevTxx := make(map[types.Hash]*Transaction)
	unspentTxx := bc.findUnspentTx(transferTx.Signer)
	for _, txIn := range transferTx.In {
		prevTx, ok := unspentTxx[txIn.TxID]
		if !ok {
			return ErrTxNotfound
		}
		prevTxx[txIn.TxID] = prevTx
	}
	if err := transferTx.Verify(prevTxx); err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) findUnspentTx(pubkey *crypto.PublicKey) map[types.Hash]*Transaction {
	transactions, err := bc.store.GetTransferState()
	if err != nil {
		panic(err)
	}
	spentTxo := map[types.Hash][]int{}
	unSpentTx := map[types.Hash]*Transaction{}

	// Loop for every transactions from the beginning
	for hash, tx := range transactions {
		transferTx := tx.TxInner.(TransferTx)
		// check for every txout in the transaction
	loop:
		for outIdx, txo := range transferTx.Out {
			//  check if txOut is existed in the spent (txInput from previous transaction)
			if spentTxo[hash] != nil {
				for _, spentOutIdx := range spentTxo[hash] {
					if outIdx == spentOutIdx {
						continue loop
					}
				}
			}

			// Check if txOut is belong to the current user
			if txo.IsLockedWithKey(pubkey) {
				unSpentTx[hash] = tx
			}
		}

		// check for every txIn in current transaction and append user's used txOut
		for _, txi := range transferTx.In {
			if txi.ScriptSig.Use(pubkey) {
				spentTxo[hash] = append(spentTxo[hash], txi.Index)
			}
		}
	}
	return unSpentTx
}

// return transactions which include txOut that could spent for amount
// total spentableTXO value >= amount
func (bc *BlockChain) FindSpentableTXO(pubKey *crypto.PublicKey, amount int) (int, map[types.Hash][]int) {
	unspentTx := bc.findUnspentTx(pubKey)
	spentableTXO := make(map[types.Hash][]int)
	accumulated := 0

loop:
	for hash, tx := range unspentTx {
		transferTx := tx.TxInner.(TransferTx)

		for outidx, txo := range transferTx.Out {
			if txo.IsLockedWithKey(pubKey) {
				accumulated += amount
			}
			spentableTXO[hash] = append(spentableTXO[hash], outidx)
			if accumulated >= amount {
				break loop
			}
		}
	}
	return accumulated, spentableTXO
}

// TODO: init blockchain with a root account?
// TODO: integrate account to blockchain
// TODO: test transfertx
