package blockchain

import (
	"blockchaintest/consts"
	"blockchaintest/drivers"
	"encoding/hex"
	"github.com/OpenStars/EtcdBackendService/StringBigsetService/bigset/thrift/gen-go/openstars/core/bigset/generic"
	"log"
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (utxo *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := map[string][]int{}

	accumulated := 0

	i := int32(0);
Work:
	for {
		it, err := drivers.GetBigsetClient().BsGetSlice(consts.UTXOBUCKET, i, 1)
		if err != nil {
			log.Panic(err, " find spendable output")
		}
		if it == nil || len(it) == 0{
			break
		}

		txID := hex.EncodeToString(it[0].GetKey())
		outs := DeserializeOutputs(it[0].GetValue())

		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated > amount {
					break Work
				}
			}
		}
		i++
	}

	return accumulated, unspentOutputs
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	UTXOs := []TXOutput{}
	i := int32(0)
	for {
		it, err := drivers.GetBigsetClient().BsGetSlice(consts.UTXOBUCKET, i, 1)
		if err != nil {
			log.Panic(err, " find spendable output")
		}
		if it == nil || len(it) == 0 {
			break
		}
		outs := DeserializeOutputs(it[0].GetValue())

		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
		i++
	}
	return UTXOs
}

func (u UTXOSet) CountTransactions() int {
	total, err := drivers.GetBigsetClient().GetTotalCount(consts.UTXOBUCKET)
	if err != nil {
		log.Panic(err, " can't count transaction")
	}
	return int(total)
}

func (u UTXOSet) Reindex() {
	_, err := drivers.GetBigsetClient().CreateStringBigSet2(consts.UTXOBUCKET)
	if err != nil {
		log.Panic(err)
	}
	for {
		it, err := drivers.GetBigsetClient().BsGetSlice2(consts.UTXOBUCKET, 0, 1)
		if err != nil {
			log.Panic(err, " error reindex")
		}
		if it == nil || len(it) == 0{
			break
		}
		_, err = drivers.GetBigsetClient().BsRemoveItem2(consts.UTXOBUCKET, it[0].GetKey())
		if err != nil {
			log.Panic(err, " error reindex")
		}
	}

	UTXO := u.Blockchain.FindUTXO()

	for txID, outs := range UTXO {
		key, err := hex.DecodeString(txID)
		if err != nil {
			log.Panic(err)
		}

		_, err = drivers.GetBigsetClient().BsPutItem2(consts.UTXOBUCKET, &generic.TItem{
			Key:   key,
			Value: outs.Serialize(),
		})
		if err != nil {
			log.Panic(err)
		}
	}
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
	for _, tx := range block.Transactions {
		if tx.IsCoinbase() == false {
			for _, vin := range tx.Vin {
				updateOuts := TXOutputs{}
				it, err := drivers.GetBigsetClient().BsGetItem2(consts.UTXOBUCKET, vin.Txid)
				if err != nil {
					log.Panic(err)
				}
				if it == nil {
					log.Panic(err, "nil item when update")
				}
				outs := DeserializeOutputs(it.GetValue())

				for outIdx, out := range outs.Outputs {
					if outIdx != vin.Vout {
						updateOuts.Outputs = append(updateOuts.Outputs, out)
					}
				}

				if len(updateOuts.Outputs) == 0 {
					_, err := drivers.GetBigsetClient().BsRemoveItem2(consts.UTXOBUCKET, vin.Txid)
					if err != nil {
						log.Panic(err)
					}
				} else {
					_, err := drivers.GetBigsetClient().BsPutItem2(consts.UTXOBUCKET, &generic.TItem{
						Key:   vin.Txid,
						Value: updateOuts.Serialize(),
					})
					if err != nil {
						log.Panic(err)
					}
				}
			}
		}

		newOutputs := TXOutputs{}
		for _, out := range tx.Vout {
			newOutputs.Outputs = append(newOutputs.Outputs, out)
		}

		_, err := drivers.GetBigsetClient().BsPutItem2(consts.UTXOBUCKET, &generic.TItem{
			Key:   tx.ID,
			Value: newOutputs.Serialize(),
		})
		if err != nil {
			log.Panic(err)
		}
	}
}