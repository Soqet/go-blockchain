package cli

import (
	"flag"
	"bchain/internal/blockchain"
	"os"
	"fmt"
	"strconv"
)

const (
	addBlockFlagName = "addblock"
	printChainFlagName = "printchain"
)

type CLI struct {
	bc *blockchain.Blockchain
}

func NewCli(bc *blockchain.Blockchain) *CLI {
	return &CLI{bc}
}

func (cli *CLI) addBlockCmd(data string) {
	fmt.Printf("Mining block %s\n", data)
	cli.bc.AddBlock([]byte(data))
	fmt.Println("Success")
}

func (cli *CLI) printChainCmd() {
	iter := cli.bc.Iterator()
	fmt.Println()
	for iter.Next() {
		block := iter.Block()
		fmt.Printf("data: %s\nhash: %x\nprevhash: %x\nnonce:%d\nvalid:%s\n\n", 
			block.Data, block.Hash, block.PrevHash, block.Nonce, strconv.FormatBool(block.Validate()),
		)
	}
}

func (cli *CLI) printHelp() {
	fmt.Println("Usage:")
	fmt.Printf("\t%s\n", addBlockFlagName)
	fmt.Printf("\t%s\n", printChainFlagName)
}

func (cli *CLI) isValidFlags() bool {
	return len(os.Args) >= 2
}

func (cli *CLI) Run() {
	if (!cli.isValidFlags()) {
		cli.printHelp()
		os.Exit(1)
	}
	addBlockFlag := flag.NewFlagSet(addBlockFlagName, flag.ExitOnError)
	addBlockData := addBlockFlag.String("data", "", "Block data")
	printChainFlag := flag.NewFlagSet(printChainFlagName, flag.ExitOnError)
	switch os.Args[1] {
	case addBlockFlagName:
		err := addBlockFlag.Parse(os.Args[2:])
		if err != nil {
			addBlockFlag.Usage()
			os.Exit(1)
		}
		cli.addBlockCmd(*addBlockData)
	case printChainFlagName:
		err := printChainFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.printChainCmd()
	default:
		cli.printHelp()
	}

}
