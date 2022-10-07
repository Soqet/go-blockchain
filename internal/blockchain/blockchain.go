package blockchain

import (
	database "bchain/internal/db"
	"bytes"
	"crypto/ecdsa"
	"errors"
)

const WalletFile = "./wallets.dat"
const BlockchainVersion = 0

type Blockchain struct {
	// last block hash
	tip []byte
	db  *database.DB
	utxoset *UTXOset
}

type BlockchainIterator struct {
	currentHash  []byte
	currentBlock *Block
	db           *database.DB
}

func NewBlockchain(db *database.DB, address string) (*Blockchain, error) {
	last, err := db.GetLast()
	if err != nil {
		return nil, err
	}
	if len(last) > 0 {
		bc := &Blockchain{tip: last, db: db}
		bc.utxoset = NewUTXOset(bc)
		return bc, nil
	}
	if address == "" {
		bc := &Blockchain{tip: []byte{}, db: db}
		bc.utxoset = NewUTXOset(bc)
		return bc, nil
	}
	coinbaseTX, err := NewCoinbaseTX(address, address)
	if err != nil {
		return nil, err
	}
	block := NewGenesisBlock(coinbaseTX)
	genesisSerialized, err := block.Serialize()
	if err != nil {
		return nil, err
	}
	err = db.UpdateLast(block.Hash)
	if err != nil {
		return nil, err
	}
	err = db.AddBlock(block.Hash, genesisSerialized)
	if err != nil {
		return nil, err
	}
	bc := &Blockchain{tip: block.Hash, db: db}
	bc.utxoset = NewUTXOset(bc)
	bc.utxoset.Reindex()
	return bc, nil
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	lastHash, err := bc.db.GetLast()
	if err != nil {
		return err
	}
	for _, tx := range transactions {
		if b, err := bc.VerifyTransaction(tx); err != nil || !b {
			return errors.New("INVALID TRANSACTION")
		}
	}
	newBlock := NewBlock(transactions, lastHash)
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
	err = bc.utxoset.UpdateWithBlock(newBlock)
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
	block, err := DeserializeBlock(serializedBlock)
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

func (bc *Blockchain) FindUnspentTX(pubKeyHash []byte) []*Transaction {
	unspentTXs := []*Transaction{}
	spentTXOs := make(map[string][]int64)
	bcIter := bc.Iterator()
	for bcIter.Next() {
		block := bcIter.Block()
		for _, tx := range block.Transactions {
		Outs:
			for i, out := range tx.Vout {
				if txo, ok := spentTXOs[string(tx.ID)]; ok {
					for _, spent := range txo {
						if spent == int64(i) {
							continue Outs
						}
					}
				}
				if out.IsLockedWith(pubKeyHash) {
					unspentTXs = append(unspentTXs, tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.IsUsesKey(pubKeyHash) {
						spentTXOs[string(in.TxID)] = append(spentTXOs[string(in.TxID)], in.Vout)
					}
				}
			}
		}
	}
	return unspentTXs
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	utxo := map[string]TXOutputs{}
	spentTXOs := make(map[string][]int64)
	bcIter := bc.Iterator()
	for bcIter.Next() {
		block := bcIter.Block()
		for _, tx := range block.Transactions {
		Outs:
			for i, out := range tx.Vout {
				if txo, ok := spentTXOs[string(tx.ID)]; ok {
					for _, spent := range txo {
						if spent == int64(i) {
							continue Outs
						}
					}
				}
				outs := utxo[string(tx.ID)]
				outs.Outputs = append(outs.Outputs, out)
				utxo[string(tx.ID)] = outs
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					spentTXOs[string(in.TxID)] = append(spentTXOs[string(in.TxID)], in.Vout)
				}
			}
		}
	}
	return utxo
}

func (bc *Blockchain) FindUnspentTXO(pubKeyHash []byte) []TXOutput {
	utxos := []TXOutput{}
	utxs := bc.FindUnspentTX(pubKeyHash)
	for _, tx := range utxs {
		for _, out := range tx.Vout {
			if out.IsLockedWith(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}
	return utxos
}

func (bc *Blockchain) NewUTXOTransaction(from string, to string, amount int64) (*Transaction, error) {
	var inputs []TXInput
	var outputs []TXOutput
	wallets, err := NewWallets(WalletFile)
	if err != nil {
		return nil, err
	}
	wallet := wallets.GetWallet(from)
	var bal int64
	var txOuts map[string][]int64
	if b, _ := bc.utxoset.IsActual(); b {
		pubKey, err := ExtractPubKeyHash(from)
		if err != nil {
			return nil, err
		}
		bal, txOuts, err = bc.utxoset.FindSpendableOuts(pubKey, amount)
		if err != nil {
			return nil, err
		}
	} else {
		bal, txOuts, err = bc.FindSpendableOuts(from, amount)
	}
	if err != nil {
		return nil, err
	}
	if bal < amount {
		return nil, errors.New("NOT ENOUGH FUNDS")
	}
	for txIDstr, outs := range txOuts {
		txId := []byte(txIDstr)
		for _, out := range outs {
			inputs = append(inputs, TXInput{TxID: txId, Vout: out, Signature: nil, PubKey: wallet.PublicKey})
		}
	}
	outputs = append(outputs, *NewTXO(amount, to))
	if bal > amount {
		outputs = append(outputs, *NewTXO(bal-amount, from))
	}
	tx := &Transaction{
		Vin:  inputs,
		Vout: outputs,
	}
	hash, err := tx.Hash()
	if err != nil {
		return nil, err
	}
	tx.ID = hash
	err = bc.SignTransaction(tx, wallet.PrivateKey)
	return tx, err
}

func (bc *Blockchain) FindSpendableOuts(from string, amount int64) (int64, map[string][]int64, error) {
	spendableOuts := map[string][]int64{}
	pubKey, err := ExtractPubKeyHash(from)
	if err != nil {
		return 0, nil, err
	}
	unspentTXs := bc.FindUnspentTX(pubKey)
	var balance int64 = 0
mainloop:
	for _, tx := range unspentTXs {
		strTxID := string(tx.ID)
		for i, out := range tx.Vout {
			if out.IsLockedWith(pubKey) {
				balance += out.Value
				spendableOuts[strTxID] = append(spendableOuts[strTxID], int64(i))
			}
			if balance >= amount {
				break mainloop
			}
		}
	}
	return balance, spendableOuts, nil
}

func (bc *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	bci := bc.Iterator()
	for bci.Next() {
		block := bci.Block()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return tx, nil
			}
		}
	}
	return nil, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TxID)
		if err != nil {
			return err
		}
		prevTXs[string(prevTX.ID)] = *prevTX
	}
	tx.Sign(privKey, prevTXs)
	return nil
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TxID)
		if err != nil {
			return false, err
		}
		prevTXs[string(prevTX.ID)] = *prevTX
	}

	return tx.Verify(prevTXs)
}
