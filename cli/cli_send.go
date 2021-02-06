package cli

import (
	"blockchaintest/blockchain"
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{bc}

	tx := blockchain.NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := blockchain.NewCoinBaseTX(from, "")
	txs := []*blockchain.Transaction{cbTx, tx}
	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Println("Success!")
}