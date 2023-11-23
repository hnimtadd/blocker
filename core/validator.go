package core

import "fmt"

type Validator interface {
	Validate(*Block) error
}

type BlockValidator struct {
	bc *BlockChain
}

func NewBlockValidator(bc *BlockChain) *BlockValidator {
	return &BlockValidator{
		bc: bc,
	}
}

func (v *BlockValidator) Validate(block *Block) error {
	if v.bc.HasBlock(block.Height) {
		return fmt.Errorf("Block (%s) with height (%d) existed", block.Hash(BlockHasher{}), block.Height)
		// panic(err)
	}

	if v.bc.Height()+1 != block.Height {
		return fmt.Errorf("Block (%s) with height (%d) => current height (%d)", block.Hash(BlockHasher{}), block.Height, v.bc.Height())
	}

	prevHeader, err := v.bc.GetHeader(uint32(block.Height - 1))
	if err != nil {
		return err
	}
	prevHash := BlockHasher{}.Hash(prevHeader)
	if prevHash != block.PrevBlockHash {
		return fmt.Errorf("Block (%s) has invalid previousDataHash(%s) => previousDataHash (%s)", block.Hash(BlockHasher{}), block.PrevBlockHash.Short(), prevHash.Short())
	}

	if err := block.Verify(); err != nil {
		return err
	}
	return nil
}
