package cli

import (
	"blockchaintest/blockchain"
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	blockchain.CreateBlockchain(address)
	fmt.Println("Done!")
}
