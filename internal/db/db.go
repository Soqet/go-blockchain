package db

import (
	"database/sql"
	// "fmt"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func NewDb(filePath string) (db *DB, err error) {
	db = new(DB)
	db.db, err = sql.Open("sqlite3", filePath)
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
	return db, nil
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