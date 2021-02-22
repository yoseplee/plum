package merkleTree

import (
	"crypto/sha256"
	"github.com/yoseplee/plum/core/plum"
	"math"
)

func NewNode(l, r *plum.MerkleNode, d []byte) *plum.MerkleNode {
	var n plum.MerkleNode
	hash := sha256.New()

	if l == nil && r == nil {
		n.L = nil
		n.R = nil
		hash.Write(d)
		n.D = hash.Sum(nil)
	} else {
		n.L = l
		n.R = r

		hash.Write(l.D)
		hash.Write(r.D)
		n.D = hash.Sum(nil)
	}

	return &n
}

func appendLast(d [][]byte) [][]byte {
	return append(d, d[len(d)-1])
}

func isEven(d [][]byte) bool {
	l := float64(len(d))
	c := math.Mod(l, 2.0)
	if c == 0 {
		return true
	} else {
		return false
	}
}

func NewTree(d [][]byte) *plum.MerkleTree {
	var t plum.MerkleTree
	var n []plum.MerkleNode

	if d == nil {
		t.Root = &plum.MerkleNode{}
		return &t
	}

	if !isEven(d) {
		d = appendLast(d)
	}

	for _, datum := range d {
		n = append(n, *NewNode(nil, nil, datum))
	}

	for {
		var level []plum.MerkleNode
		if len(n) == 1 {
			break
		}

		for i := 0; i < len(n); i += 2 {
			if len(n) == i+1 {
				level = append(level, n[i])
				continue
			}
			level = append(level, *NewNode(&n[i], &n[i+1], nil))
		}
		n = level
	}

	t.Root = &n[0]
	return &t
}
