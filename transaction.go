package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID   []byte
	VIn  []TXInput
	VOut []TXOutput
}

type TXInput struct {
	TXid      []byte
	VOut      int
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

//SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	encode := bytes.Buffer{}
	hash := [32]byte{}

	enc := gob.NewEncoder(&encode)
	err := enc.Encode(tx)
	if err != nil {
		log.Fatal(err)
	}

	hash = sha256.Sum256(encode.Bytes())
	tx.ID = hash[:]
}

//create a new coin base transaction
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprint("Reward to %s", to)
	}

	txin := TXInput{
		TXid:      []byte{},
		VOut:      -1,
		ScriptSig: data,
	}

	txout := TXOutput{
		Value:        subsity,
		ScriptPubKey: to,
	}

	tx := Transaction{
		VIn:  []TXInput{txin},
		VOut: []TXOutput{txout},
	}

	tx.SetID()

	return &tx
}

func NewUTXOTransaction(from, to string, amout int, bc *Blockchain) *Transaction {
	inputs := []TXInput{}
	outputs := []TXOutput{}

	acc, validOutputs := bc.FindSpendableOutputs(from, amout)

	if acc < amout {
		log.Panic("ERROR: not enough")
	}

	for tx, outs := range validOutputs {
		txID, err := hex.DecodeString(tx)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{
				TXid:      txID,
				VOut:      out,
				ScriptSig: from,
			}

			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{
		Value:        amout,
		ScriptPubKey: to,
	})

	if acc > amout {
		outputs = append(outputs, TXOutput{
			Value:        acc - amout,
			ScriptPubKey: from,
		})//return the change
	}

	tx := Transaction{
		ID:   nil,
		VIn:  inputs,
		VOut: outputs,
	}

	return &tx
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.VIn) == 1 && len(tx.VIn[0].TXid) == 0 && tx.VIn[0].VOut == -1
}
