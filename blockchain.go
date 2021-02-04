package main

import (
	"blockchaintest/drivers"
	"encoding/hex"
	"log"

	"github.com/OpenStars/EtcdBackendService/StringBigsetService/bigset/thrift/gen-go/openstars/core/bigset/generic"
)

type Blockchain struct {
	tip []byte
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	lastHash := []byte{}

	it, err := drivers.GetBigsetClient().BsGetItem2(LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Println(err, " blockchain.go:19")
		return
	}

	lastHash = it.GetValue()

	newBlock := NewBlock(transactions, lastHash)

	lastHash = newBlock.Hash
	_, err = drivers.GetBigsetClient().BsPutItem2(BLOCKCHAIN, &generic.TItem{
		Key:   newBlock.Hash,
		Value: newBlock.Serialize(),
	})
	if err != nil {
		log.Println(err, " blockchain.go:33")
		return
	}
	_, err = drivers.GetBigsetClient().BsPutItem2(LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: lastHash,
	})
	bc.tip = lastHash
}

//add a new block to blockchain
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	lastHash := []byte{}

	it, err := drivers.GetBigsetClient().BsGetItem2(LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Println(err, " blockchain.go:19")
		return
	}

	lastHash = it.GetValue()

	newBlock := NewBlock(transactions, lastHash)

	lastHash = newBlock.Hash
	_, err = drivers.GetBigsetClient().BsPutItem2(BLOCKCHAIN, &generic.TItem{
		Key:   newBlock.Hash,
		Value: newBlock.Serialize(),
	})
	if err != nil {
		log.Println(err, " blockchain.go:33")
		return
	}
	_, err = drivers.GetBigsetClient().BsPutItem2(LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: lastHash,
	})
	bc.tip = lastHash
}


func NewBlockchain(address string) *Blockchain {
	if !dbExists() {
		log.Fatal("Blockchain not exist")
	}

	it, err := drivers.GetBigsetClient().BsGetItem2(LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Println(err, " blockchain.go:50")
		return nil
	}

	bc := Blockchain{tip: it.GetValue()}

	return &bc
}

func dbExists() bool {
	info, err := drivers.GetBigsetClient().GetBigSetInfoByName2(BLOCKCHAIN)
	if err != nil {
		log.Println(err, " in check db exist")
		return false
	}
	if info == nil || info.Count == nil || *info.Count == -1 {
		return false
	}
	return true
}

func CreateBlockchain(address string) *Blockchain{
	if dbExists() {
		log.Fatal("Blockchain already exist")
	}

	cbtx := NewCoinBaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	_, err := drivers.GetBigsetClient().BsPutItem2(BLOCKCHAIN, &generic.TItem{
		Key:   genesis.Hash,
		Value: genesis.Serialize(),
	})

	if err != nil {
		log.Println("error putting item")
		return nil
	}

	_, err = drivers.GetBigsetClient().BsPutItem2(LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: genesis.Hash,
	})

	bc := Blockchain{tip: genesis.Hash}
	return &bc
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{currentHash: bc.tip}
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	unspentTXs := []Transaction{}
	spentTXOs := map[string][]int{}

	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs : 
			for outIdx, out := range tx.VOut {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.VIn {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TXid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.VOut)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	UTXOs := []TXOutput{}

	unspentTx := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTx {
		for _, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := map[string][]int{}
	unspentTXs := bc.FindUnspentTransactions(address)

	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}