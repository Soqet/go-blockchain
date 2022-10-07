package cli

import (
	"bchain/internal/blockchain"
	database "bchain/internal/db"
	"flag"
	"fmt"
	"os"
)

const (
	sendFlagName         = "send"
	printChainFlagName   = "printchain"
	getBalanceFlagName   = "getbalance"
	createWalletFlagName = "createwallet"
	printWalletsFlagName = "printwallets"
)

type CLI struct {
	db *database.DB
	bc *blockchain.Blockchain
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

func (cli *CLI) sendCmd(from string, to string, amount int64) {
	if b, e := blockchain.ValidateAddress(from); !b || e != nil {
		fmt.Println("ERROR: Sender address is not valid")
	}
	if b, e := blockchain.ValidateAddress(to); !b || e != nil {
		fmt.Println("ERROR: Recipient address is not valid")
	}
	err := cli.createBlockChain(from)
	if err != nil {
		fmt.Println("Something went wrong")
		return
	}
	tx, err := cli.bc.NewUTXOTransaction(from, to, amount)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cli.bc.MineBlock([]*blockchain.Transaction{tx})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Success")
}

func (cli *CLI) printChainCmd() {
	err := cli.createBlockChain("")
	if err != nil {
		fmt.Println("Something went wrong")
		return
	}
	iter := cli.bc.Iterator()
	fmt.Println()
	for iter.Next() {
		block := iter.Block()
		fmt.Printf("=================== Block %x ==================\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevHash)
		fmt.Printf("Valid: %t\n", block.Validate())
		fmt.Println("Transactions:")
		for _, tx := range block.Transactions {
			fmt.Printf("ID: %x\n", tx.ID)
			fmt.Printf("Is coinbase: %t\n", tx.IsCoinbase())
			if !tx.IsCoinbase() {
				fmt.Println("Inputs:")
				for i, txi := range tx.Vin {
					fmt.Printf("%d:\n", i)
					itx, err := cli.bc.FindTransaction(txi.TxID)
					if err != nil {
						fmt.Println("\tCANT DISPLAY INPUT")
						continue
					}
					fmt.Printf("\tValue: %d\n", itx.Vout[txi.Vout].Value)
					addr, err := blockchain.GetAddress(itx.Vout[txi.Vout].PubKeyHash, blockchain.BlockchainVersion)
					if err != nil {
						fmt.Println("\tCANT DISPLAY ADDRESS")
						continue
					}
					fmt.Printf("\tAddress: %s\n", addr)
				}
			}
			fmt.Println("Outs:")
			for i, txo := range tx.Vout {
				fmt.Printf("%d:\n", i)
				fmt.Printf("\tValue: %d\n", txo.Value)
				addr, err := blockchain.GetAddress(txo.PubKeyHash, blockchain.BlockchainVersion)
				if err != nil {
					fmt.Println("\tCANT DISPLAY ADDRESS")
					continue
				}
				fmt.Printf("\tAddress: %s\n", addr)
			}
		}
		fmt.Printf("\n\n")
	}
}

func (cli *CLI) getBalance(address string) {
	err := cli.createBlockChain(address)
	if err != nil {
		fmt.Println("Something went wrong")
		return
	}
	var balance int64 = 0
	pubKeyHash, err := blockchain.ExtractPubKeyHash(address)
	if err != nil {
		return
	}
	unspentTXOs := cli.bc.FindUnspentTXO(pubKeyHash)
	for _, utx := range unspentTXOs {
		balance += utx.Value
	}
	fmt.Printf("Balance is %d\n", balance)
}

func (cli *CLI) printHelp() {
	fmt.Println("Usage:")
	fmt.Printf("\t%s\n", sendFlagName)
	fmt.Printf("\t%s\n", printChainFlagName)
	fmt.Printf("\t%s\n", getBalanceFlagName)
	fmt.Printf("\t%s\n", createWalletFlagName)
}

func (cli *CLI) createWallet() {
	wallets, err := blockchain.NewWallets(blockchain.WalletFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	address, err := wallets.CreateWallet()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = wallets.SaveToFile(blockchain.WalletFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) printWallets() {
	wallets, err := blockchain.NewWallets(blockchain.WalletFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	for a := range wallets.Wallets {
		fmt.Println(a)
	}
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
		cli.getBalance(*getBalanceAddr)
	case createWalletFlagName:
		err := createWalletFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.createWallet()
	case printWalletsFlagName:
		err := printWalletsFlag.Parse(os.Args[2:])
		if err != nil {
			printChainFlag.Usage()
			os.Exit(1)
		}
		cli.printWallets()
	default:
		cli.printHelp()
	}

}
