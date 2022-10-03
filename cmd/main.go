package main

import (
	bchain "bchain/internal/blockchain"
	// "fmt"
	// "strconv"
	database "bchain/internal/db"
	"bchain/internal/cli"
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

	// b.AddBlock([]byte("1"))
	// b.AddBlock([]byte("2"))
	// b.AddBlock([]byte("3"))
	// b.AddBlock([]byte("4"))


	// for _, e := range b.Blocks {
	// 	fmt.Printf("data: %s\nhash: %x\nprevhash: %x\nnonce:%d\nvalid:%s\n\n", e.Data, e.Hash, e.PrevHash, e.Nonce, strconv.FormatBool(e.Validate()))
	// }
}
