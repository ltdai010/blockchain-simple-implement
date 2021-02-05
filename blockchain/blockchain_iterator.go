package blockchain

import (
	"blockchaintest/consts"
	"blockchaintest/drivers"
	"log"
)

type BlockchainIterator struct {
	currentHash []byte
}

func (bi *BlockchainIterator) Next() *Block {
	it, err := drivers.GetBigsetClient().BsGetItem2(consts.BLOCKCHAIN, bi.currentHash)
	if err != nil {
		log.Println(err, " blockchain_iterator.go:15")
		return nil
	}

	currentBlock := DeserializeBlock(it.GetValue())

	bi.currentHash = currentBlock.PrevBlockHash
	return currentBlock
}
