package merkleTree

import (
	"bytes"
	"crypto/sha256"
	"github.com/yoseplee/plum/core/plum"
	"reflect"
	"testing"
)

var (
	tData    []byte
	mempool  [][]byte
	hMempool [][]byte
)

func Test_Init(t *testing.T) {
	tData = []byte("hello world")
	mempool = [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}

	for _, datum := range mempool {
		hashed := sha256.Sum256(datum)
		hMempool = append(hMempool, hashed[:])
	}
}

func TestNewNode(t *testing.T) {
	sha := sha256.New()
	want := &plum.MerkleNode{
		L: nil,
		R: nil,
	}
	sha.Write(tData)
	data := sha.Sum(nil)
	want.D = data
	sha.Reset()

	if got := NewNode(nil, nil, tData); reflect.DeepEqual(got, want) == false {
		t.Errorf("failed to make a new terminal node\n")
	}

	want = &plum.MerkleNode{
		L: &plum.MerkleNode{L: nil, R: nil, D: sha.Sum(tData)},
		R: &plum.MerkleNode{L: nil, R: nil, D: sha.Sum(tData)},
	}
	sha.Write(want.L.D)
	sha.Write(want.R.D)
	want.D = sha.Sum(nil)
	sha.Reset()

	got := NewNode(&plum.MerkleNode{L: nil, R: nil, D: sha.Sum(tData)}, &plum.MerkleNode{L: nil, R: nil, D: sha.Sum(tData)}, tData)

	if reflect.DeepEqual(got, want) != true {
		t.Errorf("failed to make a new node\n")
		t.Errorf("want: %s\n", want.String())
		t.Errorf("got: %s\n", got.String())
	}
}

func TestAppendLast(t *testing.T) {
	want := append(mempool, []byte("e"))
	if got := appendLast(mempool); reflect.DeepEqual(got, want) == false {
		t.Errorf("failed to append the last entity")
	}
}

func TestIsEven(t *testing.T) {
	//case for odd number of entity
	want := false
	if got := isEven(mempool); got != want {
		t.Errorf("cannot decide if it has even number of entity")
	}

	//case for even number of entity
	want = true
	if got := isEven(append(mempool, []byte("f"))); got != want {
		t.Errorf("cannot decide if it has even number of entity")
	}
}

func TestNewTreeTwoNodes(t *testing.T) {
	sha := sha256.New()
	a := sha256.Sum256([]byte("a"))
	b := sha256.Sum256([]byte("b"))

	sha.Write(a[:])
	sha.Write(b[:])
	want := sha.Sum(nil)
	sha.Reset()

	tMem := [][]byte{[]byte("a"), []byte("b")}
	tree := NewTree(tMem)
	got := tree.Root.D

	if bytes.Compare(want, got) != 0 {
		t.Errorf("invalid root of the tree")
	}
}

func TestNewTreeFiveNodes(t *testing.T) {
	var l, r, e, root []byte
	sha := sha256.New()

	//level-1, (0,1)
	sha.Write(hMempool[0])
	sha.Write(hMempool[1])
	l = sha.Sum(nil)
	sha.Reset()

	//level-1, (2,3)
	sha.Write(hMempool[2])
	sha.Write(hMempool[3])
	r = sha.Sum(nil)
	sha.Reset()

	//level-1, (4,4)
	sha.Write(hMempool[4])
	sha.Write(hMempool[4])
	e = sha.Sum(nil)
	sha.Reset()

	//level-2 (0,1)
	sha.Write(l)
	sha.Write(r)
	l = sha.Sum(nil)
	sha.Reset()

	//level-2 (2) - win by default as there's no node to join
	r = e

	//level-3 (0,1) - root
	sha.Write(l)
	sha.Write(r)
	root = sha.Sum(nil)
	sha.Reset()

	want := root

	tree := NewTree(mempool)
	got := tree.Root.D

	if bytes.Compare(want, got) != 0 {
		t.Errorf("invaild root of the tree")
	}
}

func TestNewTreeNilTxs(t *testing.T) {
	if got := NewTree(nil); got == nil {
		t.Errorf("failed to make a new tree when data is empty")
	}

	if got := NewTree(nil).Root; got == nil {
		t.Errorf("failed to make a new tree when data is empty")
	}
}
