package blockchain 

import (
	"bytes"
)

type UTXOset struct {
	bc *Blockchain
}

func NewUTXOset(bc *Blockchain) *UTXOset {
	return &UTXOset{bc: bc}
}

func (uset *UTXOset) Reindex() error {
	err := uset.bc.db.ClearUTXOset()
	if err != nil {
		return err
	}
	err = uset.bc.db.UpdateUTXOBlock(uset.bc.tip)
	if err != nil {
		return err
	}
	utxo := uset.bc.FindUTXO()
	for id, txos := range utxo {
		serialized, err := txos.Serialize()
		if err != nil {
			return err
		}
		err = uset.bc.db.AddTXO([]byte(id), serialized)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uset *UTXOset) UpdateWithBlock(block *Block) error {
	for _, tx := range block.Transactions {
		newOuts := TXOutputs{}
		newOuts.Outputs = append(newOuts.Outputs, tx.Vout...)
		serialized, err := newOuts.Serialize()
		if err != nil {
			return err
		}
		err = uset.bc.db.AddTXO(tx.ID, serialized)
		if err != nil {
			return err
		}
		if tx.IsCoinbase() {
			continue
		}
		serialized, err = uset.bc.db.GetUTXO(tx.ID)
		if err != nil {
			return err
		}
		outs, err := DeserializeTXO(serialized)
		if err != nil {
			return err
		}
		updatedOuts := TXOutputs{}
		for _, txi := range tx.Vin {
			for _, out := range outs.Outputs {
				if !bytes.Equal(out.PubKeyHash, txi.PubKey) {
					updatedOuts.Outputs = append(updatedOuts.Outputs, out)
				}
			}
		}
		if len(updatedOuts.Outputs) != 0 {
			serialized, err = updatedOuts.Serialize()
			if err != nil {
				return err
			}
			err = uset.bc.db.AddTXO(tx.ID, serialized)
			if err != nil {
				return err
			}
		} else {
			err = uset.bc.db.DeleteUTXO(tx.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (uset *UTXOset) FindSpendableOuts(pubKeyHash []byte, amount int64) (int64, map[string][]int64, error) {
	spendableOuts := map[string][]int64{}
	var balance int64 = 0
	si, err := uset.bc.db.UTXOiterator()
	if err != nil {
		return 0, nil, err
	}
	for si.Next() {
		elem := si.Get()
		txId := string(elem.TxHash)
		outs, err := DeserializeTXO(elem.Txo)
		if err != nil {
			return 0, nil, err
		}
		for i, out := range outs.Outputs {
			if balance >= amount {
				return balance, spendableOuts, nil
			}
			if out.IsLockedWith(pubKeyHash) {
				balance += out.Value
				spendableOuts[txId] = append(spendableOuts[txId], int64(i))
			}
		}
	}
	return balance, spendableOuts, nil 
}

func (uset *UTXOset) FindUnspentTXO(pubKeyHash []byte) ([]TXOutput, error) {
	utxos := []TXOutput{}
	si, err := uset.bc.db.UTXOiterator()
	if err != nil {
		return nil, err
	}
	for si.Next() {
		elem := si.Get()
		outs, err := DeserializeTXO(elem.Txo)
		if err != nil {
			return nil, err
		}
		for _, out := range outs.Outputs {
			if out.IsLockedWith(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}
	return utxos, nil
}