package db

import (
	"database/sql"
)

type Iterator[T any] interface {
	Next() bool
	// must call next before calling get
	Get() T
}

type TXOiterator struct {
	rows *sql.Rows
}

type UTXOsetElem struct {
	TxHash []byte
	Txo    []byte
}

type KnownNodesIterator struct {
	rows *sql.Rows
}

type KnownNodesElem struct {
	Address string
	Version int32
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

func (iter *KnownNodesIterator) Next() bool {
	if !iter.rows.Next() {
		iter.rows.Close()
		return false
	}
	return true
}

func (iter *KnownNodesIterator) Get() KnownNodesElem {
	res := KnownNodesElem{}
	iter.rows.Scan(&res.Address, &res.Version)
	return res
}
