package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"math/big"
	"strconv"
	"time"
)

type Block struct {
	Version      uint32
	Timestamp    int64
	Transactions []*Transaction
	Hash         []byte
	PrevHash     []byte
	Nbits        uint8
	Nonce        uint64
}

func NewBlock(transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Version:      BlockchainVersion,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		Hash:         []byte{},
		PrevHash:     prevHash,
		Nbits:        16,
	}
	pow := NewPoW(block)
	pow.RunParallel()
	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) GetData() []byte {
	return b.getDataNonce(b.Nonce)
}

func (b *Block) getDataNonce(nonce uint64) []byte {
	data := bytes.Join(
		[][]byte{
			[]byte(strconv.FormatUint(uint64(b.Version), 16)),
			[]byte(strconv.FormatUint(uint64(b.Timestamp), 16)),
			b.HashTransctions(),
			b.PrevHash,
			[]byte(strconv.FormatUint(uint64(b.Nbits), 16)),
			[]byte(strconv.FormatUint(uint64(nonce), 16)),
		},
		[]byte{},
	)
	return data
}

func (b *Block) HashTransctions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func (b *Block) Validate() bool {
	data := b.GetData()
	hash := sha256.Sum256(data)
	var hashNum big.Int
	hashNum.SetBytes(hash[:])
	return hashNum.Cmp(getTarget(b.Nbits)) == -1
}

func (b *Block) Serialize() ([]byte, error) {
	result := bytes.Buffer{}
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		return []byte{}, err
	}
	return result.Bytes(), nil
}

func DeserializeBlock(data []byte) (*Block, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	block := new(Block)
	err := decoder.Decode(block)
	if err != nil {
		return block, err
	}
	return block, nil
}
