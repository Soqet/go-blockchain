package blockchain

type Blockchain struct {
	Blocks []Block
}

func NewBlockchain() *Blockchain {
	bchain := &Blockchain{[]Block{}}
	bchain.Blocks = append(bchain.Blocks, *NewBlock([]byte("genesis block"), []byte{}))
	return bchain
}

func (bc *Blockchain) AddBlock(data []byte) {
	newBlock := NewBlock(data, bc.Blocks[len(bc.Blocks)-1].Hash)
	bc.Blocks = append(bc.Blocks, *newBlock)
}
