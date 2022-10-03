package blockchain

import (
	"crypto/sha256"
	"math"
	"math/big"
	"sync"
)

type PoW struct {
	block  *Block
	target *big.Int
}

func getTarget(nBits uint8) *big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-int(nBits)))
	return target
}

func NewPoW(b *Block) *PoW {
	return &PoW{b, getTarget(b.Nbits)}
}

func (pow *PoW) GetData() []byte {
	return pow.block.GetData()
}

func (pow *PoW) Run() (hash [32]byte) {
	for pow.block.Nonce < math.MaxUint64 {
		data := pow.GetData()
		hash = sha256.Sum256(data)
		var hashNum big.Int
		hashNum.SetBytes(hash[:])
		if hashNum.Cmp(pow.target) == -1 {
			pow.block.Hash = hash[:]
			return
		}
		pow.block.Nonce++
	}
	return
}

func (pow *PoW) RunParallel() [32]byte {
	const computeSize = 32
	computeArr := [computeSize][32]byte{}
	for pow.block.Nonce < math.MaxUint64 {
		wg := sync.WaitGroup{}
		for i := range computeArr {
			if !(uint64(i)+pow.block.Nonce < math.MaxUint64) {
				break
			}
			wg.Add(1)
			go func(delta uint64) {
				defer wg.Done()
				data := pow.block.getDataNonce(pow.block.Nonce + uint64(delta))
				computeArr[delta] = sha256.Sum256(data)

			}(uint64(i))
		}
		wg.Wait()
		for i := range computeArr {
			var hashNum big.Int
			if hashNum.SetBytes(computeArr[i][:]); hashNum.Cmp(pow.target) == -1 {
				pow.block.Hash = computeArr[i][:]
				pow.block.Nonce += uint64(i)
				return computeArr[i]
			}
		}
		pow.block.Nonce += computeSize
	}
	return [32]byte{}
}
