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
			hash TEXT UNIQUE,
			block TEXT
		)`,
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) AddBlock(hash []byte, block []byte) error {
	_, err := db.db.Exec("REPLACE INTO blocks ( hash, block ) VALUES ( $1, $2 )", string(hash), string(block))
	return err
}

func (db *DB) UpdateLast(hash []byte) error {
	_, err := db.db.Exec("REPLACE INTO blocks ( hash, block ) VALUES ( $1, $2 )", string(hash), "l")
	return err
}

func (db *DB) GetBlock(hash []byte) ([]byte, error) {
	rows, err := db.db.Query("SELECT block FROM blocks WHERE hash = $1", hash)
	if err != nil {
		return []byte{}, nil
	}
	defer rows.Close()
	for rows.Next() {
		var block string
		rows.Scan(&block)
		return []byte(block), nil
	}
	return []byte{}, nil
}

func (db *DB) GetLast() ([]byte, error) {
	rows, err := db.db.Query("SELECT hash FROM blocks WHERE block = $1", "l")
	if err != nil {
		return []byte{}, nil
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		rows.Scan(&hash)
		return []byte(hash), nil
	}
	return []byte{}, nil
}