package blockchain

import (
	"crypto/sha256"
	"errors"
)

type MerkleTree struct {
	Root *MerkleNode
}

type MerkleNode struct {
	Data  []byte
	Left  *MerkleNode
	Right *MerkleNode
}

// non hashed data, panics if data has nil
func NewMerkleTree(hashedData [][]byte) *MerkleTree {
	var nodes []*MerkleNode
	if len(hashedData)%2 != 0 {
		hashedData = append(hashedData, hashedData[len(hashedData)-1])
	}
	for _, nodeData := range hashedData {
		// newNode, _ := NewMerkleNode(nil, nil, nodeData)
		newNode := NewMerkleLeaf(nodeData)
		nodes = append(nodes, newNode)
	}
	for i := 0; i < len(hashedData)/2; i++ {
		var newLevel []*MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node, err := NewMerkleNode(nodes[j], nodes[j+1], nil)
			if err != nil {
				panic(err)
			}
			newLevel = append(newLevel, node)
		}
		nodes = newLevel
	}
	tree := MerkleTree{Root: nodes[0]}
	return &tree
}

func NewMerkleNode(left, right *MerkleNode, data []byte) (*MerkleNode, error) {
	newNode := &MerkleNode{Left: left, Right: right}
	var hash [32]byte
	if newNode.IsLeaf() {
		hash = sha256.Sum256(data)
	} else if left == nil || right == nil {
		return nil, errors.New("INCORRECT DATA FOR NEW MERKLE NODE")
	} else {
		hash = sha256.Sum256(append(left.Data, right.Data...))
	}
	newNode.Data = hash[:]
	return newNode, nil
}

func NewMerkleLeaf(hashedData []byte) *MerkleNode {
	return &MerkleNode{Data: hashedData, Right: nil, Left: nil}
}

func (node *MerkleNode) IsLeaf() bool {
	return node.Left == nil && node.Right == nil
}
