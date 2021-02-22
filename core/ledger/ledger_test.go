package ledger

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/util/path"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"testing"
)

var l *Ledger

func TestMain(m *testing.M) {
	l = NewLedger(path.GetInstance().LedgerPath, path.GetInstance().GenesisBlockPath, true)
	exitVal := m.Run()
	flushBlocks()
	os.Exit(exitVal)
}

func generateTx() [][]byte {
	var txs [][]byte
	for i := 0; i < 2000; i++ {
		txs = append(txs, []byte(fmt.Sprintf("tx%d", rand.Intn(100000))))
	}
	return txs
}

func flushBlocks() {
	deleteCmd := fmt.Sprintf("%s %s%s", "yes y | rm", os.ExpandEnv("$PLUM_ROOT"), "/ledger_store/*")
	_, err := exec.Command("/bin/sh", "-c", deleteCmd).Output()
	if err != nil {
		log.Fatalf("could not delete all the blocks: %v", err)
	}
}

func TestLedger_String(t *testing.T) {
	log.Println(l.String())
}

func TestLedger_Append(t *testing.T) {
	maxBlock := 100
	for i := 1; i < maxBlock; i++ {
		ph := block.Digest(l.CurrentBlockHeader())
		b := block.NewBlock(generateTx(), ph, l.Height+1)
		err := l.Append(b)
		if err != nil {
			t.Errorf("could not append properly: %v", err)
		}
	}
	log.Println(l.String())
}

func BenchmarkLedger_Append(b *testing.B) {
	el := NewLedger(path.GetInstance().LedgerPath, path.GetInstance().GenesisBlockPath, false)
	ph := block.Digest(l.CurrentBlockHeader())
	nb := block.NewBlock(generateTx(), ph, el.Height+1)
	err := el.Append(nb)
	if err != nil {
		b.Errorf("could not append properly: %v", err)
	}
}

func TestLedger_LoadHeaders(t *testing.T) {
	el := NewLedger(path.GetInstance().LedgerPath, path.GetInstance().GenesisBlockPath, false)
	el.SetHeaders(el.LoadHeaders())
	for i, h := range l.Headers {
		lmh, err := proto.Marshal(h)
		if err != nil {
			t.Errorf("error on marshal message")
		}
		elmh, err := proto.Marshal(el.Headers[i])
		if err != nil {
			t.Errorf("error on marshal message")
		}
		if bytes.Compare(lmh, elmh) != 0 {
			t.Errorf("invalid loading of headers from file")
		}
	}
}

func TestLoadGenesisBlock(t *testing.T) {
	gb := LoadGenesisBlock(path.GetInstance().GenesisBlockPath)
	if gb.GetHeader().GetId() != 0 {
		t.Errorf("invalid genesis block")
	}
}
