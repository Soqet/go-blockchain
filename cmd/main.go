package main

import (
	bchain "bchain/internal/blockchain"
	"bchain/internal/cli"
	database "bchain/internal/db"
)

func main() {
	db, err := database.NewDb("./blocks.db")
	if err != nil {
		panic(db)
	}
	bc, err := bchain.NewBlockchain(db)
	if err != nil {
		panic(err)
	}
	cli.NewCli(bc).Run()
}
