package blockchain

import (
	"blockchaintest/consts"
	"blockchaintest/drivers"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"github.com/OpenStars/EtcdBackendService/StringBigsetService/bigset/thrift/gen-go/openstars/core/bigset/generic"
	"log"
)

type Blockchain struct {
	tip []byte
}

const genesisCoinbaseData = "This blockchain created by Diaz, belong to everyone"

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	lastHash := []byte{}

	it, err := drivers.GetBigsetClient().BsGetItem2(consts.LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Println(err, " blockchain.go:19")
		return
	}

	lastHash = it.GetValue()

	newBlock := NewBlock(transactions, lastHash)

	lastHash = newBlock.Hash
	_, err = drivers.GetBigsetClient().BsPutItem2(consts.BLOCKCHAIN, &generic.TItem{
		Key:   newBlock.Hash,
		Value: newBlock.Serialize(),
	})
	if err != nil {
		log.Println(err, " blockchain.go:33")
		return
	}
	_, err = drivers.GetBigsetClient().BsPutItem2(consts.LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: lastHash,
	})
	bc.tip = lastHash
}

//add a new block to blockchain
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	lastHash := []byte{}

	it, err := drivers.GetBigsetClient().BsGetItem2(consts.LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Panic(err, " blockchain.go:19")
		return nil
	}

	lastHash = it.GetValue()

	newBlock := NewBlock(transactions, lastHash)

	lastHash = newBlock.Hash
	_, err = drivers.GetBigsetClient().BsPutItem2(consts.BLOCKCHAIN, &generic.TItem{
		Key:   newBlock.Hash,
		Value: newBlock.Serialize(),
	})
	if err != nil {
		log.Panic(err, " blockchain.go:33")
		return nil
	}
	_, err = drivers.GetBigsetClient().BsPutItem2(consts.LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: lastHash,
	})

	if err != nil {
		log.Panic(err, " cannot put last hash")
	}
	bc.tip = lastHash

	return newBlock
}


func NewBlockchain() *Blockchain {
	if !dbExists() {
		log.Fatal("Blockchain not exist")
	}

	it, err := drivers.GetBigsetClient().BsGetItem2(consts.LASTHASH, generic.TItemKey("l"))
	if err != nil {
		log.Println(err, " blockchain.go:50")
		return nil
	}

	bc := Blockchain{tip: it.GetValue()}

	return &bc
}

func dbExists() bool {
	info, err := drivers.GetBigsetClient().GetBigSetInfoByName2(consts.BLOCKCHAIN)
	if err != nil {
		log.Println(err, " in check db exist")
		return false
	}
	if info == nil || info.Count == nil || *info.Count == -1 {
		return false
	}
	return true
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		log.Fatal("Blockchain already exist")
	}

	cbtx := NewCoinBaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)

	_, err := drivers.GetBigsetClient().BsPutItem2(consts.BLOCKCHAIN, &generic.TItem{
		Key:   genesis.Hash,
		Value: genesis.Serialize(),
	})

	if err != nil {
		log.Println("error putting item")
		return nil
	}

	_, err = drivers.GetBigsetClient().BsPutItem2(consts.LASTHASH, &generic.TItem{
		Key:   []byte("l"),
		Value: genesis.Hash,
	})

	bc := Blockchain{tip: genesis.Hash}
	return &bc
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{currentHash: bc.tip}
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := map[string]Transaction{}

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err, " err verify transaction")
		}
		prevTxs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTxs)
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := map[string]Transaction{}

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err, " error sign transaction")
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXOs := map[string]TXOutputs{}
	spentTXOs := map[string][]int{}
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXOs
}
