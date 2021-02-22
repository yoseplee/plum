package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"log"
	"math/rand"
	"testing"
)

func generateTx() [][]byte {
	var txs [][]byte
	for i := 0; i < 1000; i++ {
		txs = append(txs, []byte(fmt.Sprintf("tx%d", rand.Intn(100000))))
	}
	return txs
}

func TestDigest(t *testing.T) {
	fmt.Println(hex.EncodeToString(Digest(&plum.Header{})))
}

func TestNewBlock(t *testing.T) {
	txs := [][]byte{[]byte("a"), []byte("b")}
	tempPrevHash := sha256.Sum256([]byte("a"))
	b := NewBlock(txs, tempPrevHash[:], 5)
	fmt.Println(b.String())
}

func TestNewBlock2(t *testing.T) {
	txs := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	tempPrevHash := sha256.Sum256([]byte("a"))
	b := NewBlock(txs, tempPrevHash[:], 5)
	fmt.Println(b.String())
}

func TestNewBlock_5_blocks_chained(t *testing.T) {
	b0 := NewBlock(generateTx(), nil, 0) //genesis block
	b1 := NewBlock(generateTx(), Digest(b0.Header), 1)
	b2 := NewBlock(generateTx(), Digest(b1.Header), 2)
	b3 := NewBlock(generateTx(), Digest(b2.Header), 3)
	b4 := NewBlock(generateTx(), Digest(b3.Header), 4)
	b5 := NewBlock(generateTx(), Digest(b4.Header), 5) //final block

	blocks := []*plum.Block{b0, b1, b2, b3, b4, b5}
	//verify
	for i := 4; i > 1; i-- {
		cbd := blocks[i].Header.PrevBlockHash
		pbd := Digest(blocks[i-1].Header)
		if bytes.Compare(cbd, pbd[:]) != 0 {
			cbds := hex.EncodeToString(cbd)
			pbds := hex.EncodeToString(pbd[:])
			t.Errorf("invalid prev hash. prev hash: %v, actual hash of prev block: %v\n", cbds, pbds)
		}
	}
}

func TestNewBlock_Chained_1000_times(t *testing.T) {
	var blocks []plum.Block
	maxBlock := 1000

	gb := NewGenesisBlock()
	prevHash := Digest(gb.Header)
	for i := 0; i < maxBlock; i++ {
		b := NewBlock(generateTx(), prevHash, uint64(i+1))
		blocks = append(blocks, *b)
		prevHash = Digest(b.Header)
	}

	for i := maxBlock - 1; i > 1; i-- {
		cbd := blocks[i].Header.PrevBlockHash
		pbd := Digest(blocks[i-1].Header)
		if bytes.Compare(cbd, pbd[:]) != 0 {
			cbds := hex.EncodeToString(cbd)
			pbds := hex.EncodeToString(pbd[:])
			t.Errorf("invalid prev hash. prev hash: %v, actual hash of prev block: %v\n", cbds, pbds)
		}
	}
}

func BenchmarkNewBlock(b *testing.B) {
	b0 := NewBlock(generateTx(), nil, 0) //genesis block
	NewBlock(generateTx(), Digest(b0.Header), 1)
}

func TestNewGenesisBlock(t *testing.T) {
	g := NewGenesisBlock()
	log.Println(g.String())
}

func TestCompareBlock(t *testing.T) {
	a := NewBlock(nil, nil, 0)
	b := a
	if got := CompareBlock(a, b); got == false {
		t.Errorf("invalid comparison: should be true")
	}

	a = NewBlock(nil, nil, 0)
	b = NewBlock(nil, nil, 1) // time is different
	if got := CompareBlock(a, b); got == true {
		t.Errorf("invalid comparison: should be false")
	}
}

func TestCompareBlockDigest(t *testing.T) {
	a := NewBlock(nil, nil, 0)
	b := a
	if got := CompareBlockDigest(Digest(a.Header), Digest(b.Header)); got == false {
		t.Errorf("invalid comparison: should be true")
	}

	a = NewBlock(nil, nil, 0)
	b = NewBlock(nil, nil, 1) // time is different
	if got := CompareBlockDigest(Digest(a.Header), Digest(b.Header)); got == true {
		t.Errorf("invalid comparison: should be false")
	}
}
