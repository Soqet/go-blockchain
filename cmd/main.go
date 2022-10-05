package main

import (
	"bchain/internal/cli"
	database "bchain/internal/db"
)

func main() {
	db, err := database.NewDb("./blocks.db")
	if err != nil {
		panic(db)
	}
	cli.NewCli(db).Run()
}
