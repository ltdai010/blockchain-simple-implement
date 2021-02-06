package merkletree

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil || right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHash)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	nodes := []MerkleNode{}

	//if data has odd number of transaction, add one
	if len(data)%2 != 0 {
		data = append(data, data[len(data) - 1])
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for {
		if len(nodes) == 1 {
			break
		}
		newLevel := []MerkleNode{}

		for j := 0; j < len(nodes); j+= 2 {
			node := &MerkleNode{}
			if j == len(nodes) - 1 {
				node = NewMerkleNode(&nodes[j], &nodes[j], nil)
			} else {
				node = NewMerkleNode(&nodes[j], &nodes[j + 1], nil)
			}
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	return &MerkleTree{&nodes[0]}
}
