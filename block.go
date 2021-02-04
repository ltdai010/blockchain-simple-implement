package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//make a Hash of the block
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.HashTransaction(), timestamp}, []byte{})

	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

//Hash all Transaction in block
func (b *Block) HashTransaction() []byte {
	txHashes := [][]byte{}
	txHash := [32]byte{}

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

//serialize block to byte
func (b *Block) Serialize() []byte {
	result := bytes.Buffer{}
	encoder := gob.NewEncoder(&result)

	_ = encoder.Encode(b)

	return result.Bytes()
}

//deserialize byte to block
func DeserializeBlock(d []byte) *Block {
	block := Block{}
	decoder := gob.NewDecoder(bytes.NewReader(d))

	_ = decoder.Decode(&block)

	return &block
}

//create a new block with Data and prev block, set a Hash for it
func NewBlock(tx []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  tx,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}
	pow := NewProofOfWork(block)

	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash[:]

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}
