package peer

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yoseplee/plum/core/ledger"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/peer/heap"
	"github.com/yoseplee/plum/core/peer/mq"
	"github.com/yoseplee/plum/core/plum"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

var p *peer
var s *server
var conn *grpc.ClientConn

func TestMain(m *testing.M) {
	defer conn.Close()
	Setup()
	go s.Run()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func Setup() {
	//setup for peer
	p = GetInstance()
	profile := loadProfile()
	p.Init(0, "localhost", ":50051", profile, "PBFT")
	k = GetKeeperInstance()
	peerSetup()

	//setup for server
	s = NewServer()
	s.RegisterServers()
	connectClientForServerTest()
}

func peerSetup() {
	go p.D.Run()
	log.Println("connect to all peers in the address book after 1 sec")
	<-time.After(time.Second)
	p.connectAll()
	p.generateAndSetKeyPair()
	setAddressBookForTest()
}

//for test setup, insert public keys into the address boo
//mimicking sendPublicKeyToAll()
func setAddressBookForTest() {
	for i, a := range p.AddressBook {
		if i == 0 {
			continue
		}
		pk, _, _ := ed25519.GenerateKey(nil)
		a.PublicKey = pk
	}
}

func loadProfile() map[uint32]*Connection {
	profile := make(map[uint32]*Connection)

	profile[0] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50051",
		PeerId: 0,
	}

	profile[1] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50061",
		PeerId: 1,
	}

	profile[2] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50071",
		PeerId: 2,
	}

	profile[3] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50081",
		PeerId: 3,
	}

	profile[4] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50091",
		PeerId: 4,
	}

	profile[5] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50101",
		PeerId: 5,
	}

	profile[6] = &Connection{
		Ipv4:   "localhost",
		Port:   ":50111",
		PeerId: 6,
	}

	return profile
}

func TestDealer_Schedule(t *testing.T) {
	d := &Dealer{
		MQ:                  mq.NewPBFTQueue(),
		ReservedPBFTMessage: heap.NewMinPBFTHeap(),
		stopSig:             make(chan struct{}),
	}
	//mock for d.Run
	go func() {
		for {
			select {
			case <-d.stopSig:
				log.Println("Delaer stopped")
				return
			default:
				_, err := d.MQ.Pop()
				if err != nil {
					break
				}
			}
			_, err := d.ReservedPBFTMessage.Pop()
			if err != nil {
				continue
			}
		}
	}()

	//add to mq
	go func() {
		for i := 0; i < 10000; i++ {
			d.MQ.Push(&plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i % 3),
					Round: uint64(i),
				},
			})
		}
	}()

	go func() {
		for i := 0; i < 12000; i++ {
			d.ReservedPBFTMessage.Push(&plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i % 3),
					Round: uint64(i),
				},
			})
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func() {
		<-time.After(time.Millisecond * 10)
		for {
			if d.MQ.GetN() == 0 && d.ReservedPBFTMessage.Last == -1 {
				cancel()
			}
		}
	}()

	select {
	case <-ctx.Done():
		close(d.stopSig)
		mqWant := uint64(0)
		rWant := -1
		mqGot := d.MQ.GetN()
		rGot := d.ReservedPBFTMessage.Last
		if mqWant != mqGot || rWant != rGot {
			t.Errorf("the dealer didn't accomplishe all the task. mqGot: %v, mqWant: %v, rGot: %v, rWant: %v", mqGot, mqWant, rGot, rWant)
		}
	}
}

func TestPeer_Init_Ledger(t *testing.T) {
	if genesisId := p.L.Genesis.Header.Id; genesisId != 0 {
		t.Errorf("invalid ledger initiated on peer")
	}

	if initialHeight := p.L.Height; initialHeight != 0 {
		t.Errorf("invalid ledger height initiated on peer")
	}
}

func TestPeer_NewPrimary(t *testing.T) {
	var nr uint64
	var want uint32
	testSet := []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	answerSet := []uint32{0, 1, 2, 3, 4, 5, 6, 0, 1, 2, 3, 4, 5, 6, 0, 1, 2, 3, 4, 5, 6}

	nr = 1
	want = 1
	if got := p.NewPrimary(nr); want != got {
		t.Errorf("invalid calculation of new primary for new round(%d). got := %v, want := %v", nr, got, want)
		log.Println(p.String())
	}

	nr = 2
	want = 2
	if got := p.NewPrimary(nr); want != got {
		t.Errorf("invalid calculation of new primary for new round(%d). got := %v, want := %v", nr, got, want)
	}

	nr = 3
	want = 3
	if got := p.NewPrimary(nr); want != got {
		t.Errorf("invalid calculation of new primary for new round(%d). got := %v, want := %v", nr, got, want)
	}

	nr = 4
	want = 4
	if got := p.NewPrimary(nr); want != got {
		t.Errorf("invalid calculation of new primary for new round(%d). got := %v, want := %v", nr, got, want)
	}

	for i, ts := range testSet {
		want = answerSet[i]
		if got := p.NewPrimary(ts); want != got {
			t.Errorf("invalid calculation of new primary for new round(%d). got := %v, want := %v", nr, got, want)
		}
	}
}

func TestGetInstance(t *testing.T) {
	var iSlice []*peer
	for i := 0; i < 100; i++ {
		iSlice = append(iSlice, GetInstance())
	}

	//verify - change the value of attribute for 100 times and see if it is kept safely
	var lastRn uint32
	for i := 0; i < 100; i++ {
		rn := rand.Uint32() % 10
		tp := GetInstance()
		tp.ID = rn
		lastRn = rn
	}

	for _, o := range iSlice {
		if o.ID != lastRn {
			t.Errorf("failed to keep the one instance on struct peer - different ID value")
		}
	}
}

func TestPeer_CreateSignature(t *testing.T) {
	p := GetInstance()
	b := block.NewBlock(nil, nil, 0)
	bd, err := proto.Marshal(b)
	if err != nil {
		log.Printf("could not marshal the message: %v\n", err)
	}
	sig1 := p.CreateSignature(b)
	if got := ed25519.Verify(p.PublicKey, bd, sig1); got != true {
		t.Errorf("invalid verify of signature")
	}

	b = ledger.LoadGenesisBlock("../")
	bd, err = proto.Marshal(b)
	if err != nil {
		log.Printf("could not marshal the message: %v\n", err)
	}
	sig2 := p.CreateSignature(b)
	if got := ed25519.Verify(p.PublicKey, bd, sig2); got != true {
		t.Errorf("invalid verify of signature")
	}

	if got := ed25519.Verify(p.PublicKey, bd, sig1); got != false {
		t.Errorf("invalid verify of signature")
	}

	var cm *plum.PBFTMessage
	var sig []byte
	p = GetInstance()

	cm = &plum.PBFTMessage{
		Phase:  0,
		Round:  0,
		Digest: nil,
		PeerId: 0,
	}
	sig = p.CreateSignature(cm)
	cmd, err := proto.Marshal(cm)

	if err != nil {
		t.Errorf("could not marshal the message: %v\n", err)
	}

	if got := ed25519.Verify(p.PublicKey, cmd, sig); got != true {
		t.Errorf("invalid verify of signature on req")
	}
}

func TestPeer_VerifyRequestSignature(t *testing.T) {
	var cm *plum.PBFTMessage
	var req *plum.PBFTRequest
	var sig []byte
	p = GetInstance()

	cm = &plum.PBFTMessage{
		Phase:  0,
		Round:  0,
		Digest: nil,
		PeerId: 0,
	}
	sig = p.CreateSignature(cm)
	req = &plum.PBFTRequest{
		Message:   cm,
		Signature: sig,
		Block:     nil,
	}

	if got := p.VerifyConsensusMessageSignature(req); got != true {
		t.Errorf("invalid verify on request signature")
	}

	cm = &plum.PBFTMessage{
		Phase:  1,
		Round:  2,
		Digest: nil,
		PeerId: 0,
	}
	sig = p.CreateSignature(cm)
	req = &plum.PBFTRequest{
		Message:   cm,
		Signature: sig,
		Block:     nil,
	}

	if got := p.VerifyConsensusMessageSignature(req); got != true {
		t.Errorf("invalid verify on request signature")
	}

	cm = &plum.PBFTMessage{
		Phase:  1,
		Round:  2,
		Digest: nil,
		PeerId: 2,
	}
	sig = p.CreateSignature(cm)
	req = &plum.PBFTRequest{
		Message:   cm,
		Signature: sig,
		Block:     nil,
	}

	if got := p.VerifyConsensusMessageSignature(req); got != false {
		t.Errorf("invalid verify on request signature")
	}
}

func BenchmarkPeer_RetrieveTxs_Parallel(b *testing.B) {
	var txs [][]byte
	var wg sync.WaitGroup

	nGoroutines := 4
	for iter := 0; iter < nGoroutines; iter++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//var partialTx [][]byte
			for i := 0; i < 500; i++ {
				txs = append(txs, []byte(fmt.Sprintf("tx%d", rand.Intn(100000))))
			}
			//partialTxChan <- partialTx
		}()
	}
	wg.Wait()
}

func BenchmarkPeer_RetrieveTxs(b *testing.B) {
	p.RetrieveTxs()
}

func TestPeer_toleranceBase(t *testing.T) {
	var want int
	var tp *peer
	tp = GetInstance()

	want = 2
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f). got: %v, want: %v", got, want)
	}

	var dummyAddressBook map[uint32]*Connection
	dummyAddressBook = make(map[uint32]*Connection)
	for i := 0; i < 7; i++ {
		dummyAddressBook[uint32(i)] = &Connection{}
	}
	tp = &peer{AddressBook: dummyAddressBook}

	want = 2
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f)")
	}

	dummyAddressBook = make(map[uint32]*Connection)
	for i := 0; i < 10; i++ {
		dummyAddressBook[uint32(i)] = &Connection{}
	}
	tp = &peer{AddressBook: dummyAddressBook}

	want = 3
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f)")
	}

	dummyAddressBook = make(map[uint32]*Connection)
	for i := 0; i < 13; i++ {
		dummyAddressBook[uint32(i)] = &Connection{}
	}
	tp = &peer{AddressBook: dummyAddressBook}

	want = 4
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f)")
	}

	dummyAddressBook = make(map[uint32]*Connection)
	for i := 0; i < 16; i++ {
		dummyAddressBook[uint32(i)] = &Connection{}
	}
	tp = &peer{AddressBook: dummyAddressBook}

	want = 5
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f)")
	}

	dummyAddressBook = make(map[uint32]*Connection)
	for i := 0; i < 19; i++ {
		dummyAddressBook[uint32(i)] = &Connection{}
	}
	tp = &peer{AddressBook: dummyAddressBook}

	want = 6
	if got := tp.toleranceBase(); got != want {
		t.Errorf("invalid calculation of tolerance base(f)")
	}
}

func TestDealer_discardAllTheRemainedMessagesAtTheRound(t *testing.T) {
	targetCount := 0
	maxIter := 100

	//clear out heap before the test
	close(p.D.stopSig)
	p.ConsensusRound = 0
	for i := 0; i < maxIter; i++ {
		round := uint64(rand.Intn(100))
		if round == p.ConsensusRound {
			targetCount++
		}
		p.D.ReservedPBFTMessage.Push(&plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: round,
			},
		})
	}

	p.D.discardAllTheRemainedMessagesAtTheRound()
	var want int
	want = (maxIter - 1) - targetCount

	if got := p.D.ReservedPBFTMessage.Last; got != want {
		t.Errorf("mismatch between heap last and counted last: got: %v, want: %v", got, want)
	}

	want = maxIter - targetCount
	if got := len(p.D.ReservedPBFTMessage.Q); got != want {
		t.Errorf("mismatch between heap len and counted len: got: %v, want: %v", got, want)
	}

	//initialize again as this test case modified peer state
	Setup()
}

func TestPeer_EmptyAddressBook(t *testing.T) {
	var want bool

	//normal case: address book is not empty
	want = false
	if got := p.EmptyAddressBook(); got != want {
		t.Errorf("address book is not empty. got: %v, want: %v", got, want)
	}

	//exceptional case: when address book is empty
	p.AddressBook = make(map[uint32]*Connection)
	want = true
	if got := p.EmptyAddressBook(); got == want {
		t.Errorf("address book is empty. got: %v, want: %v", got, want)
	}
	log.Println("AFTER")
	Setup()
}

func Test_setXBFTThreshold(t *testing.T) {
	for i := 0; i < 4; i++ {
		p.D.committeeMembers = append(p.D.committeeMembers, &plum.CommitteeMembers{})
	}
	p.setXBFTThreshold(p.ID, p.D.committeeMembers)
	var wantPrepareThreshold int
	var wantCommitThreshold int

	wantPrepareThreshold = 1
	wantCommitThreshold = 2

	if got := p.XBFTThreshold[p.ID][plum.XBFTPhase_XBFTPrepare]; got != wantPrepareThreshold {
		t.Errorf("invalid calculation of prepare threshold when peer quantity is 4 under the same reputation. got: %v, want: %v", got, wantPrepareThreshold)
	}
	if got := p.XBFTThreshold[p.ID][plum.XBFTPhase_XBFTCommit]; got != wantCommitThreshold {
		t.Errorf("invalid calculation of commit threshold when peer quantity is 4 under the same reputation. got: %v, want: %v", got, wantCommitThreshold)
	}
}

func Test_findCommitteeMemberById(t *testing.T) {
	var cms []*plum.CommitteeMembers
	cms = append(cms, &plum.CommitteeMembers{PeerId: 0, SelectionValue: 0.0})
	cms = append(cms, &plum.CommitteeMembers{PeerId: 1, SelectionValue: 0.1})
	cms = append(cms, &plum.CommitteeMembers{PeerId: 2, SelectionValue: 0.2})
	cms = append(cms, &plum.CommitteeMembers{PeerId: 3, SelectionValue: 0.3})
	cms = append(cms, &plum.CommitteeMembers{PeerId: 4, SelectionValue: 0.4})

	var cm *plum.CommitteeMembers
	var err error
	var want float64

	cm, err = findCommitteeMemberById(0, cms)
	want = 0.0
	if got := cm.SelectionValue; got != want && err != nil {
		t.Errorf("invalid searching of committee members. got:%v, want:%v", got, want)
	}

	cm, err = findCommitteeMemberById(1, cms)
	want = 0.1
	if got := cm.SelectionValue; got != want && err != nil {
		t.Errorf("invalid searching of committee members. got:%v, want:%v", got, want)
	}

	cm, err = findCommitteeMemberById(2, cms)
	want = 0.2
	if got := cm.SelectionValue; got != want && err != nil {
		t.Errorf("invalid searching of committee members. got:%v, want:%v", got, want)
	}

	cm, err = findCommitteeMemberById(3, cms)
	want = 0.3
	if got := cm.SelectionValue; got != want && err != nil {
		t.Errorf("invalid searching of committee members. got:%v, want:%v", got, want)
	}

	cm, err = findCommitteeMemberById(4, cms)
	want = 0.4
	if got := cm.SelectionValue; got != want && err != nil {
		t.Errorf("invalid searching of committee members. got:%v, want:%v", got, want)
	}
}

func TestPeer_makeCert(t *testing.T) {
	//add cert at the first time
	//this caused an error that the nested map was not initialized properly
	p.D.CandidateBlockCertificates = make(map[uint32]map[plum.XBFTPhase][]*plum.XBFTRequest)
	makeCert(0, plum.XBFTPhase_XBFTPrepare, &plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrepare,
		},
	})
}
