package db

import (
	"database/sql"
	"errors"
	// "fmt"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

type Iterator [T any] interface {
	Next() bool
	// must call next before calling get
	Get() T
}

type TXOiterator struct {
	rows *sql.Rows
}

type UTXOsetElem struct {
	TxHash []byte
	Txo []byte
}

func NewTXOiterator() {

}

func (iter *TXOiterator) Next() bool {
	if !iter.rows.Next() {
		iter.rows.Close()
		return false
	}
	return true
}

func (iter *TXOiterator) Get() UTXOsetElem {
	res := UTXOsetElem{}
	iter.rows.Scan(&res.TxHash, &res.Txo)
	return res
}


func NewDb(blocksPath string) (db *DB, err error) {
	db = new(DB)
	db.db, err = sql.Open("sqlite3", blocksPath)
	if err != nil {
		return nil, err
	}
	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS blocks ( 
			hash BLOB UNIQUE,
			block BLOB
		)`,
	)
	if err != nil {
		return nil, err
	}
	_, err = db.db.Exec(`
		CREATE TABLE IF NOT EXISTS utxoset ( 
			hash BLOB UNIQUE,
			utxo BLOB
		)`,
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) ClearUTXOset() error {
	_, err :=db.db.Exec("DROP TABLE IF EXISTS utxoset")
	if err != nil {
		return err
	}
	_, err = db.db.Exec(`
	CREATE TABLE IF NOT EXISTS utxoset ( 
		hash BLOB UNIQUE,
		utxo BLOB
	)`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddBlock(hash []byte, block []byte) error {
	_, err := db.db.Exec("REPLACE INTO blocks ( hash, block ) VALUES ( $1, $2 )", hash, block)
	return err
}

func (db *DB) UpdateLast(hash []byte) error {
	_, err := db.db.Exec("REPLACE INTO blocks ( hash, block ) VALUES ( $1, $2 )", "l", hash)
	return err
}

func (db *DB) GetBlock(hash []byte) ([]byte, error) {
	rows, err := db.db.Query("SELECT block FROM blocks WHERE hash = $1", hash)
	if err != nil {
		return []byte{}, nil
	}
	defer rows.Close()
	for rows.Next() {
		var block []byte
		rows.Scan(&block)
		return block, nil
	}
	return []byte{}, nil
}

func (db *DB) GetLast() ([]byte, error) {
	rows, err := db.db.Query("SELECT block FROM blocks WHERE hash = $1", "l")
	if err != nil {
		return []byte{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash []byte
		rows.Scan(&hash)
		return hash, nil
	}
	return []byte{}, nil
}

func (db *DB) AddTXO(hash []byte, txo []byte) error {
	_, err := db.db.Exec("REPLACE INTO utxoset ( hash, utxo ) VALUES ( $1, $2 )", hash, txo)
	return err
}

func (db *DB) UpdateUTXOBlock(blockHash []byte) error {
	_, err := db.db.Exec("REPLACE INTO utxoset ( hash, utxo ) VALUES ( $1, $2 )", "b", blockHash)
	return err
}

func (db *DB) GetUTXO(transactionHash []byte) ([]byte, error) {
	rows, err := db.db.Query("SELECT utxo FROM utxoset WHERE hash = $1", transactionHash)
	if err != nil {
		return []byte{}, nil
	}
	defer rows.Close()
	for rows.Next() {
		var block []byte
		rows.Scan(&block)
		return block, nil
	}
	return []byte{}, nil
}

func (db *DB) GetUTXOBlock() ([]byte, error) {
	rows, err := db.db.Query("SELECT utxo FROM utxoset WHERE hash = $1", "b")
	if err != nil {
		return []byte{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash []byte
		rows.Scan(&hash)
		return hash, nil
	}
	return []byte{}, nil
}

func (db *DB) DeleteUTXO(hash []byte) (error) {
	if string(hash) == "b" {
		return errors.New("INVALID TX HASH")
	}
	_, err := db.db.Exec("DELETE FROM utxoset WHERE hash = $1", hash)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UTXOiterator() (Iterator[UTXOsetElem], error) {
	rows, err := db.db.Query("SELECT ( hash, utxo ) FROM utxoset WHERE hash != $1", "b")
	if err != nil {
		return nil, err
	}
	return &TXOiterator{rows: rows}, nil
}
