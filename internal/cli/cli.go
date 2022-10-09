package cli

import (
	"bchain/internal/blockchain"
	database "bchain/internal/db"
	"flag"
	"os"
)

const (
	sendFlagName         = "send"
	printChainFlagName   = "printchain"
	getBalanceFlagName   = "balance"
	createWalletFlagName = "createwallet"
	printWalletsFlagName = "printwallets"
	listenFlagName       = "listen"
	helpFlagName         = "help"
)

type CLI struct {
	db      *database.DB
	bc      *blockchain.Blockchain
	utxoSet *blockchain.UTXOset
}

func NewCli(db *database.DB) *CLI {
	return &CLI{
		db: db,
	}
}

func (cli *CLI) createBlockChain(address string) error {
	bc, err := blockchain.NewBlockchain(cli.db, address)
	if err != nil {
		return err
	}
	cli.bc = bc
	utxoset := blockchain.NewUTXOset(bc)
	err = utxoset.Reindex()
	if err != nil {
		return err
	}
	cli.utxoSet = utxoset
	return nil
}

func (cli *CLI) isValidFlags() bool {
	return len(os.Args) >= 2
}

func (cli *CLI) Run() {
	if !cli.isValidFlags() {
		cli.printHelp()
		os.Exit(1)
	}

	sendFlag := flag.NewFlagSet(sendFlagName, flag.ExitOnError)
	sendFrom := sendFlag.String("f", "", "from addres")
	sendTo := sendFlag.String("t", "", " to address")
	sendAmount := sendFlag.Int64("a", 0, "amount")

	printChainFlag := flag.NewFlagSet(printChainFlagName, flag.ExitOnError)

	getBalanceFlag := flag.NewFlagSet(getBalanceFlagName, flag.ExitOnError)
	getBalanceAddr := getBalanceFlag.String("a", "", "address")

	createWalletFlag := flag.NewFlagSet(createWalletFlagName, flag.ExitOnError)

	printWalletsFlag := flag.NewFlagSet(printWalletsFlagName, flag.ExitOnError)

	listenFlag := flag.NewFlagSet(listenFlagName, flag.ExitOnError)
	listenAddr := listenFlag.String("a", "", "address")

	switch os.Args[1] {
	case sendFlagName:
		err := sendFlag.Parse(os.Args[2:])
		if err != nil {
			sendFlag.Usage()
			os.Exit(1)
		}
		cli.sendCmd(*sendFrom, *sendTo, *sendAmount)
	case printChainFlagName:
		err := printChainFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.printChainCmd()
	case getBalanceFlagName:
		err := getBalanceFlag.Parse(os.Args[2:])
		if err != nil {
			getBalanceFlag.Usage()
			os.Exit(1)
		}
		cli.getBalanceCmd(*getBalanceAddr)
	case createWalletFlagName:
		err := createWalletFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.createWalletCmd()
	case printWalletsFlagName:
		err := printWalletsFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.printWalletsCmd()
	case listenFlagName:
		err := listenFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.listenCmd(*listenAddr)
	case helpFlagName:
		fallthrough
	default:
		cli.printHelp()
	}

}
