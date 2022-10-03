package blockchain

import (
	database "bchain/internal/db"
)

type Blockchain struct {
	tip []byte
	db  *database.DB
}

type BlockchainIterator struct {
	currentHash  []byte
	currentBlock *Block
	db           *database.DB
}

func NewBlockchain(db *database.DB) (*Blockchain, error) {
	last, err := db.GetLast()
	if err != nil {
		return nil, err
	}
	if len(last) > 0 {
		return &Blockchain{last, db}, nil
	}
	block := NewBlock([]byte("genesis block"), []byte{}) // genesis block
	genesisHash, err := block.Serialize()
	if err != nil {
		return nil, err
	}
	err = db.UpdateLast(block.Hash)
	if err != nil {
		return nil, err
	}
	err = db.AddBlock(block.Hash, genesisHash)
	if err != nil {
		return nil, err
	}
	return &Blockchain{genesisHash, db}, nil
}

func (bc *Blockchain) AddBlock(data []byte) error {
	lastHash, err := bc.db.GetLast()
	if err != nil {
		return err
	}
	newBlock := NewBlock(data, lastHash)
	err = bc.db.UpdateLast(newBlock.Hash)
	if err != nil {
		return err
	}
	newSerialized, err := newBlock.Serialize()
	if err != nil {
		return err
	}
	err = bc.db.AddBlock(newBlock.Hash, newSerialized)
	if err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, nil, bc.db}
}

func (bci *BlockchainIterator) Next() bool {
	if len(bci.currentHash) == 0 {
		return false
	}
	serializedBlock, err := bci.db.GetBlock(bci.currentHash)
	if err != nil {
		return false
	}
	block, err := Deserialize(serializedBlock)
	if err != nil {
		return false
	}
	bci.currentHash = block.PrevHash
	bci.currentBlock = block
	return true
}

func (bci *BlockchainIterator) Block() *Block {
	return bci.currentBlock
}
