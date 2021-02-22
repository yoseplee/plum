package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/plum"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Ledger struct {
	Genesis    *plum.Block
	Headers    []*plum.Header
	Height     uint64
	path       string
	storeBlock bool
}

//NewLedger is to create a new ledger. lp stands for ledger path and gbp stands for genesis block path, therefore the two parameter should be path to them.
//note that genesis block path should exclude filename which will be added in runtime.
//example: NesLedger("~/", "~/genesis/")
func NewLedger(lp, gbp string, storeBlock bool) *Ledger {
	var ledgerPath string
	ledgerPath = adjustLedgerPath(lp)

	//1. make a genesis block
	//this can be replaced to receiving it from other peer or file
	gb := LoadGenesisBlock(gbp)
	l := &Ledger{
		Genesis:    gb,
		Height:     0,
		path:       ledgerPath,
		storeBlock: storeBlock,
	}
	l.Headers = append(l.Headers, gb.Header)
	l.toFile(gb)

	return l
}

func adjustGenesisBlockPath(gbp string) string {
	var p string

	if gbp == "" {
		p = fmt.Sprintf("%s%s", os.ExpandEnv("$PLUM_ROOT"), "/core/genesis.block")
	} else {
		p = gbp
	}

	if !strings.HasSuffix(gbp, "genesis.block") {
		p = fmt.Sprintf("%sgenesis.block", gbp)
	}

	return p
}

func adjustLedgerPath(lp string) string {
	var p string

	if lp == "" {
		p = fmt.Sprintf("%s%s", os.ExpandEnv("$PLUM_ROOT"), "/ledger_store/")
	} else {
		p = lp
	}

	//path suffix check - it should include '/'
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func (l *Ledger) String() string {
	var s string
	s += fmt.Sprintf("Printing Ledger\n")
	s += fmt.Sprintf("Height: %d\n", l.Height)
	return s
}

func (l *Ledger) Append(b *plum.Block) error {
	//1) verify prev hash
	bh := block.Digest(l.CurrentBlockHeader())
	if bytes.Compare(bh, b.Header.PrevBlockHash) != 0 {
		return errors.New("the block has different previous block hash against the ledger")
	}

	//2) keep header into the ledger
	l.Headers = append(l.Headers, b.Header)
	l.Height++

	//3) save entire block into file system: Marshal/Unmarshal is needed
	l.toFile(b)
	return nil
}

func (l *Ledger) toFile(b *plum.Block) {
	//if ledger options is not to store block, skip
	if l.storeBlock == false {
		return
	}

	m, mErr := proto.Marshal(b)
	if mErr != nil {
		log.Println("could not marshall the message")
	}

	err := ioutil.WriteFile(
		fmt.Sprintf("%sblock-%d.block", l.path, b.Header.Id),
		m,
		os.FileMode(777),
	)

	if err != nil {
		log.Println("could not store the block")
		return
	}
}

func (l *Ledger) CurrentBlockHeader() *plum.Header {
	return l.Headers[l.Height]
}

func (l *Ledger) SetHeaders(h []*plum.Header) {
	l.Headers = h
	newHeight := uint64(len(h) - 1)
	if newHeight < 0 {
		newHeight = 0
	}
	l.Height = newHeight
}

func (l *Ledger) LoadHeaders() []*plum.Header {
	var headers []*plum.Header
	log.Println("path: ", l.path)
	var i uint64
	for {
		rb, err := ioutil.ReadFile(fmt.Sprintf("%sblock-%d.block", l.path, i))
		if err != nil {
			log.Println("could not read file")
			break
		}
		b := &plum.Block{}
		mErr := proto.Unmarshal(rb, b)
		if mErr != nil {
			log.Fatalf("could not unmarshal properly: %d th block\n", i)
		}
		headers = append(headers, b.Header)
		i++
	}
	return headers
}

func LoadGenesisBlock(gbp string) *plum.Block {
	genesisBlockPath := adjustGenesisBlockPath(gbp)

	rb, fErr := ioutil.ReadFile(genesisBlockPath)
	if fErr != nil {
		log.Fatalf("could not read genesis block properly: %v", fErr)
	}
	gb := &plum.Block{}
	mErr := proto.Unmarshal(rb, gb)
	if mErr != nil {
		log.Fatalf("could not unmarshal properly: %v", mErr)
	}
	return gb
}

func (l *Ledger) GetBlockById(id uint64) (*plum.Block, error) {
	return nil, nil
}

func (l *Ledger) GetBlockAll() []*plum.Block { return nil }
