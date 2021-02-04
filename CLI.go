package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {}

func (cli *CLI) createBlockchain(address string) {
	CreateBlockchain(address)
	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet(GETBALANCE, flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet(CREATEBLOCKCHAIN, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(SEND, flag.ExitOnError)
	printBlockCmd := flag.NewFlagSet(PRINTCHAIN, flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String(ADDRESS, "", "The address to get balance")
	createBlockchainAddress := createBlockchainCmd.String(ADDRESS, "", "The address to send genesis" +
		" block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case GETBALANCE:
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Println(err, " CLI.go:23")
		}
	case CREATEBLOCKCHAIN:
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case PRINTCHAIN:
		err := printBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Println(err, " CLI.go:29")
		}
	case SEND:
		err := sendCmd.Parse(os.Args[2:])
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

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bc := NewBlockchain("")

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}


func main() {
	cli := CLI{}
	cli.Run()
}
