package cli

import (
	"bchain/internal/blockchain"
	"fmt"
)

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
	coinbaseTx, err := blockchain.NewCoinbaseTX(from, "")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cli.bc.MineBlock([]*blockchain.Transaction{coinbaseTx, tx})
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
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Version: %d\n", block.Version)
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

func (cli *CLI) getBalanceCmd(address string) {
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
	fmt.Println("Commands:")

	fmt.Printf("\t%s\n", getBalanceFlagName)
	fmt.Printf("\t\tUsage: %s -a <address>\n", getBalanceFlagName)

	fmt.Printf("\t%s\n", sendFlagName)
	fmt.Printf("\t\tUsage: %s -a <address from> -t <address to> -a <amount>\n", sendFlagName)

	fmt.Printf("\t%s\n", printChainFlagName)
	fmt.Printf("\t\tUsage: %s\n", printChainFlagName)

	fmt.Printf("\t%s\n", createWalletFlagName)
	fmt.Printf("\t\tUsage: %s\n", createWalletFlagName)

	fmt.Printf("\t%s\n", printWalletsFlagName)
	fmt.Printf("\t\tUsage: %s\n", printChainFlagName)

	fmt.Printf("\t%s\n", listenFlagName)
	fmt.Printf("\t\tUsage: %s -a <address>\n", listenFlagName)
}

func (cli *CLI) createWalletCmd() {
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

func (cli *CLI) printWalletsCmd() {
	wallets, err := blockchain.NewWallets(blockchain.WalletFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	for a := range wallets.Wallets {
		fmt.Println(a)
	}
}

func (cli *CLI) listenCmd(address string) {
	err := cli.createBlockChain(address)
	if err != nil {
		fmt.Println("Something went wrong")
		return
	}
	fmt.Println("Not implemented")
}
