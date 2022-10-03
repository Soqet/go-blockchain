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
	Version   uint32
	Timestamp int64
	Data      []byte
	Hash      []byte
	PrevHash  []byte
	Nbits     uint8
	Nonce     uint64
}

func (b *Block) SetHash() {
	time := []byte(strconv.FormatInt(b.Timestamp, 10))
	concat := bytes.Join([][]byte{b.PrevHash, b.Data, time}, []byte{})
	hash := sha256.Sum256(concat)
	b.Hash = hash[:]
}

func NewBlock(data []byte, prevHash []byte) *Block {
	block := &Block{
		Version:   0,
		Timestamp: time.Now().Unix(),
		Data:      data,
		Hash:      []byte{},
		PrevHash:  prevHash,
		Nbits:     16,
	}
	pow := NewPoW(block)
	pow.RunParallel()
	return block
}

func (b *Block) GetData() []byte {
	return b.getDataNonce(b.Nonce)
}

func (b *Block) getDataNonce(nonce uint64) []byte {
	data := bytes.Join(
		[][]byte{
			[]byte(strconv.FormatUint(uint64(b.Version), 16)),
			[]byte(strconv.FormatUint(uint64(b.Timestamp), 16)),
			b.Data,
			b.PrevHash,
			[]byte(strconv.FormatUint(uint64(b.Nbits), 16)),
			[]byte(strconv.FormatUint(uint64(nonce), 16)),
		},
		[]byte{},
	)
	return data
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

func Deserialize(data []byte) (*Block, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	block := new(Block)
	err := decoder.Decode(block)
	if err != nil {
		// fmt.Printf("%s\n%s", string(data), err.Error())
		return block, err
	}
	return block, nil
}
