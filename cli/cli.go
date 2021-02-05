package cli

import (
	"blockchaintest/consts"
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {}


func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet(consts.GETBALANCE, flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet(consts.CREATEBLOCKCHAIN, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(consts.SEND, flag.ExitOnError)
	printBlockCmd := flag.NewFlagSet(consts.PRINTCHAIN, flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet(consts.CREATEWALLET, flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet(consts.LISTADDRESS, flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String(consts.ADDRESS, "", "The address to get balance")
	createBlockchainAddress := createBlockchainCmd.String(consts.ADDRESS, "", "The address to send genesis" +
		" block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case consts.GETBALANCE:
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Println(err, " CLI.go:23")
		}
	case consts.CREATEBLOCKCHAIN:
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case consts.PRINTCHAIN:
		err := printBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Println(err, " CLI.go:29")
		}
	case consts.SEND:
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case consts.CREATEWALLET:
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case consts.LISTADDRESS:
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}

		cli.getBalance(*getBalanceAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printBlockCmd.Parsed() {
		cli.printChain()
	}
	return
}


