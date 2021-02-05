package blockchain

import (
	"blockchaintest/utils"
	"bytes"
	"crypto/sha256"
	"log"
	"math"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

const targetBits = 24

const maxNonce = math.MaxInt64

//create a new proof of work
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256 -targetBits))

	pow := &ProofOfWork{
		block:  b,
		target: target,
	}

	return pow
}

//prepare Data
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{pow.block.PrevBlockHash, pow.block.HashTransaction(), utils.IntToHex(pow.block.Timestamp),
		utils.IntToHex(int64(targetBits)), utils.IntToHex(int64(nonce))}, []byte{})

	return data
}

//start mining
func (pow *ProofOfWork) Run() (int, []byte) {
	hashInt := big.Int{}
	hash := [32]byte{}
	nonce := 0

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)

		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	log.Println("finish mining block")
	return nonce, hash[:]
}

//validate proof of work
func (pow *ProofOfWork) Validate() bool {
	hashInt := big.Int{}

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
