package block

import (
	"bytes"
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/yoseplee/plum/core/ledger/merkleTree"
	"github.com/yoseplee/plum/core/plum"
	"log"
	"time"
)

//Digest hashes a block into array of byte.
func Digest(bh *plum.Header) []byte {
	m, err := proto.Marshal(bh)
	if err != nil {
		log.Println("could not marshall the message")
	}
	d := sha256.Sum256(m)
	return d[:]
}

func NewBlock(txs [][]byte, prevBlockHash []byte, id uint64) *plum.Block {
	mTree := merkleTree.NewTree(txs)
	//assign
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		log.Println("could not convert timestamp for protobuf")
	}

	b := plum.Block{
		Header: &plum.Header{
			Id:            id,
			MerkleRoot:    mTree.Root.D,
			PrevBlockHash: prevBlockHash,
			Time:          timestamp,
		},
		Body: &plum.Body{
			MerkleTree: mTree,
			Txs:        txs,
		},
	}

	return &b
}

func NewGenesisBlock() *plum.Block {
	var g *plum.Block
	g = NewBlock(nil, nil, 0)
	return g
}

func CompareBlock(a, b *plum.Block) bool {
	//marshal and compare
	am, amErr := proto.Marshal(a)
	if amErr != nil {
		log.Fatalf("could not marshal the message: %v", a)
	}
	bm, bmErr := proto.Marshal(b)
	if bmErr != nil {
		log.Fatalf("could not marshal the message: %v", b)
	}
	if bytes.Compare(am, bm) != 0 {
		return false
	}
	return true
}

func CompareBlockDigest(a, b []byte) bool {
	if bytes.Compare(a, b) != 0 {
		log.Println("failed to compare digests of the two block")
		return false
	}
	return true
}
