package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
)

const reward = 50

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	TxID      []byte
	Vout      int64
	Signature []byte
	PubKey    []byte
}

type TXOutput struct {
	Value      int64
	PubKeyHash []byte
}

func NewCoinbaseTX(to string, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("coinbase to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXO(reward, to)
	tx, err := NewTX([]TXInput{txin}, []TXOutput{*txout})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func NewTX(vin []TXInput, vout []TXOutput) (*Transaction, error) {
	tx := &Transaction{Vin: vin, Vout: vout}
	hash, err := tx.Hash()
	if err != nil {
		return nil, err
	}
	tx.ID = hash
	return tx, nil
}

func (tx Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.TxID, vin.Vout, nil, nil})
	}
	outputs := append([]TXOutput{}, tx.Vout...)
	return Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: outputs,
	}
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}
	txCopy := tx.TrimmedCopy()
	for i, vin := range txCopy.Vin {
		prevTX := prevTXs[string(vin.TxID)]
		txCopy.Vin[i].Signature = nil
		txCopy.Vin[i].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		id, err := txCopy.Hash()
		if err != nil {
			return err
		}
		txCopy.ID = id
		txCopy.Vin[i].PubKey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return err
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[i].Signature = signature
	}
	return nil
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[string(vin.TxID)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		id, err := txCopy.Hash()
		if err != nil {
			return false, err
		}
		txCopy.ID = id
		txCopy.Vin[inID].PubKey = nil
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false, nil
		}
	}
	return true, nil
}

func (tx Transaction) IsCoinbase() bool {
	if len(tx.Vin) > 1 {
		return false
	}
	return tx.Vin[0].Vout == -1
}

func (tx *Transaction) Serialize() ([]byte, error) {
	encoded := new(bytes.Buffer)
	encoder := gob.NewEncoder(encoded)
	err := encoder.Encode(tx)
	if err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}

func (tx Transaction) Hash() ([]byte, error) {
	idBuf := bytes.Buffer{}
	encoder := gob.NewEncoder(&idBuf)
	tx.ID = []byte{}
	err := encoder.Encode(tx)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(idBuf.Bytes())
	return hash[:], err
}

func (in *TXInput) IsUsesKey(keyHash []byte) bool {
	hash := sha256.Sum256(in.PubKey)
	return bytes.Equal(keyHash, hash[:])
}

func (out *TXOutput) Lock(address string) error {
	pubKeyHash, err := ExtractPubKeyHash(address)
	if err != nil {
		return err
	}
	out.PubKeyHash = pubKeyHash
	return nil
}

func (out *TXOutput) IsLockedWith(keyHash []byte) bool {
	return bytes.Equal(keyHash, out.PubKeyHash)
}

func NewTXO(value int64, address string) *TXOutput {
	txo := TXOutput{value, nil}
	txo.Lock((address))
	return &txo
}
