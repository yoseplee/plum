package peer

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yoseplee/plum/core/ledger"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/peer/heap"
	"github.com/yoseplee/plum/core/peer/messageLog"
	"github.com/yoseplee/plum/core/peer/mq"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"github.com/yoseplee/plum/core/util/path"
	"google.golang.org/grpc"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

type peer struct {
	ID                     uint32
	Primary                uint32
	XBFTPrimary            uint32
	PrivateKey             ed25519.PrivateKey
	PublicKey              ed25519.PublicKey
	Role                   plum.ConsensusRole
	Malicious              bool
	ConsensusRound         uint64
	SelectedCount          uint64
	TentativeSelectedCount uint64
	PBFTPhase              plum.PBFTPhase
	XBFTPhase              plum.XBFTPhase
	PBFTVote               map[plum.PBFTPhase]int
	ConsensusState         plum.ConsensusState
	PBFTThreshold          map[plum.PBFTPhase]int
	XBFTThreshold          map[uint32]map[plum.XBFTPhase]int
	Ipv4                   string
	Port                   string
	AddressBook            map[uint32]*Connection
	ReputationBook         map[uint32]float64
	D                      *Dealer
	K                      *Keeper
	MQ                     *mq.Queue
	XBFTMQ                 *mq.XQueue
	ReservedPBFTMessage    *heap.MinPBFTHeap
	ReservedXBFTMessage    *heap.MinXBFTHeap
	XBFTMessageLog         messageLog.LogManager
	L                      *ledger.Ledger
	rwMutex                *sync.RWMutex
	mutex                  *sync.Mutex
}

type Connection struct {
	PeerId          uint32
	ContainerName   string
	PublicKey       ed25519.PublicKey
	Ipv4            string
	Port            string
	clientConn      *grpc.ClientConn
	consensusClient plum.ConsensusClient
	peerClient      plum.PeerClient
}

var (
	once     sync.Once
	instance *peer
)

func (p *peer) Init(id uint32, ipv4 string, port string, profile map[uint32]*Connection, consensusType string) {
	p.ID = id

	// [EXPERIMENT] set malicious nodes
	if id == 0 || id == 1 {
		p.Malicious = true
	}

	p.Primary = 0

	p.Role = plum.ConsensusRole_Backup
	p.PBFTPhase = plum.PBFTPhase_PBFTNewRound
	p.XBFTPhase = plum.XBFTPhase_XBFTPrePrepare
	p.AddressBook = profile
	p.ReputationBook = make(map[uint32]float64)
	p.initReputation()
	p.MQ = mq.NewPBFTQueue()
	p.XBFTMQ = mq.NewXBFTQueue()
	p.XBFTMessageLog = &messageLog.MessageLog{}
	p.D = NewDealer(consensusType)
	p.K = GetKeeperInstance()
	p.L = ledger.NewLedger(path.GetInstance().LedgerPath, path.GetInstance().GenesisBlockPath, false)

	p.XBFTThreshold = make(map[uint32]map[plum.XBFTPhase]int)
	p.Ipv4 = ipv4
	p.Port = port
	p.PBFTVote = make(map[plum.PBFTPhase]int)
	p.setPBFTThreshold()
	p.ConsensusState = plum.ConsensusState_Idle
	p.rwMutex = &sync.RWMutex{}
	p.mutex = &sync.Mutex{}
}

func (p *peer) run() {
	log.Println("connect to all peers in the address book after 3 sec")
	<-time.After(time.Second * 3)
	p.connectAll()
	p.generateAndSetKeyPair()

	setPubKeyDone := make(chan struct{})
	go func() {
		p.sendPublicKeyToAll()
		close(setPubKeyDone)
	}()

	go func() {
		<-setPubKeyDone
		go p.D.Run()
		p.triggerConsensus()
	}()
}

func (p *peer) InitAndRun(id uint32, ipv4 string, port string, profile map[uint32]*Connection, consensusType string) {
	p.Init(id, ipv4, port, profile, consensusType)
	p.run()
}

//triggerConsensus runs only if the very first time of consensus, when peer id is 0 and all the public key is set in this peer
func (p *peer) triggerConsensus() {

	log.Println("ready to consensus!")
	if p.ID != 0 {
		return
	}

	log.Println("trigger consensus after 5 seconds")
	<-time.After(time.Second * 5)

	if p.EmptyAddressBook() {
		return
	}

	log.Println("try to trigger consensus for 15 seconds")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	log.Println("start to trigger...")

	//trigger: send PBFTRequest to the peer itself
	nextBlock := p.NewCandidateBlock()
	nextBlockDigest, pErr := proto.Marshal(nextBlock)
	if pErr != nil {
		log.Fatalf("could not make candidate block: %v", pErr)
	}

	switch p.D.ConsensusType {
	case "PBFT":
		consensusMessage := &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTNewRound,
			Round:  0,
			Digest: nextBlockDigest,
			PeerId: 0,
		}

		sig := p.CreateSignature(consensusMessage)

		_, err := p.AddressBook[0].consensusClient.ServePBFTPhase(ctx, &plum.PBFTRequest{
			Message:   consensusMessage,
			Signature: sig,
			Block:     nextBlock,
		}, grpc.WaitForReady(true))

		if err != nil {
			log.Fatalf("could not trigger the consensus: %v", err)
		}

		log.Println("PBFT consensus triggered!")
	case "XBFT":
		cms := []*plum.CommitteeMembers{
			{
				PeerId:         0,
				Round:          0,
				SelectionValue: 1.3,
			},
			{
				PeerId:         1,
				Round:          0,
				SelectionValue: 0.9,
			},
			{
				PeerId:         2,
				Round:          0,
				SelectionValue: 0.95,
			},
			{
				PeerId:         3,
				Round:          0,
				SelectionValue: 0.94,
			},
		}
		nextBlock.CommitteeMembers = cms

		consensusMessage := &plum.XBFTMessage{
			Phase:  plum.XBFTPhase_XBFTPrePrepare,
			Round:  0,
			Height: p.L.Height,
			Digest: nextBlockDigest,
			PeerId: 0,
		}

		sig := p.CreateSignature(consensusMessage)

		req := &plum.XBFTRequest{
			Message:   consensusMessage,
			Signature: sig,
			Block:     nextBlock,
		}
		go SendAll(req)

		log.Println("XBFT consensus triggered!")
	}
}

func (p *peer) EmptyAddressBook() bool {
	for _, a := range p.AddressBook {
		if len(a.PublicKey) == 0 || len(a.PublicKey) != ed25519.PublicKeySize || a.PublicKey == nil {
			return true
		}
	}
	return false
}

func (p *peer) generateAndSetKeyPair() {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("could not generate key pair: %v", err)
	}
	p.PrivateKey = privateKey
	p.PublicKey = publicKey
	p.AddressBook[p.ID].PublicKey = p.PublicKey
}

func (p *peer) sendPublicKeyToAll() {
	log.Println("broadcast public key to all for 10 seconds")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var wg sync.WaitGroup
	for _, conn := range p.AddressBook {
		wg.Add(1)
		go func(conn *Connection) {
			defer wg.Done()
			_, err := conn.peerClient.SetPublicKey(ctx, &plum.PublicKey{
				Id:   p.ID,
				Ipv4: p.Ipv4,
				Port: p.Port,
				Key:  p.PublicKey,
			}, grpc.WaitForReady(true))

			if err != nil {
				log.Printf("could not set public key: %v", err)
			}
		}(conn)
	}
	wg.Wait()
}

//setPBFTThreshold sets consensus threshold according to size of the address book
func (p *peer) setPBFTThreshold() {
	p.PBFTThreshold = make(map[plum.PBFTPhase]int)
	//as the thresholds are compared without '>' operator instead of '>=', minus -1 is applied here
	p.PBFTThreshold[plum.PBFTPhase_PBFTPrepare] = (p.toleranceBase() * 2) - 1 //2f
	p.PBFTThreshold[plum.PBFTPhase_PBFTCommit] = p.toleranceBase() * 2        //2f+1
	p.PBFTThreshold[plum.PBFTPhase_PBFTRoundChange] = p.toleranceBase() * 2   //2f+1
}

//setXBFTThreshold sets consensus threshold according to size of the address book
func (p *peer) setXBFTThreshold(peerID uint32, committeeMembers []*plum.CommitteeMembers) {
	//as the thresholds are compared without '>' operator instead of '>=', minus -1 is applied here
	if p.XBFTThreshold[peerID] == nil {
		p.XBFTThreshold[peerID] = make(map[plum.XBFTPhase]int)
	}
	p.XBFTThreshold[peerID][plum.XBFTPhase_XBFTPrepare] = int(math.Floor((p.xbftToleranceBase(committeeMembers) * 2) - 1)) //2f
	p.XBFTThreshold[peerID][plum.XBFTPhase_XBFTCommit] = int(math.Floor(p.xbftToleranceBase(committeeMembers) * 2))        //2f+1
	p.XBFTThreshold[peerID][plum.XBFTPhase_XBFTRoundChange] = int(math.Floor(p.xbftToleranceBase(committeeMembers) * 2))   //2f+1
}

func (p *peer) toleranceBase() int {
	return int(math.Floor(float64(len(p.AddressBook)-1) / 3))
}

func (p *peer) xbftToleranceBase(committeeMembers []*plum.CommitteeMembers) float64 {
	return float64(len(committeeMembers)-1) / 3.0
}

func (p peer) String() string {
	var s string
	switch p.D.ConsensusType {
	case "PBFT":
		p.mutex.Lock()
		t, err := util.MakeTarget(p.Ipv4, p.Port)
		if err != nil {
			log.Println("could not make target, ", err)
		}
		s += fmt.Sprintf("\n| %s |\n", "================ PEER ================")
		s += fmt.Sprintf("|%-17s| %-20d |\n", "ID", p.ID)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Address", t)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Role", p.Role.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Round", p.ConsensusRound)
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Current Primary", p.Primary)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus Phase", p.PBFTPhase.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[Prepare]", p.PBFTVote[plum.PBFTPhase_PBFTPrepare])
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[Commit]", p.PBFTVote[plum.PBFTPhase_PBFTCommit])
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[RoundChange]", p.PBFTVote[plum.PBFTPhase_PBFTRoundChange])
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus State", p.ConsensusState.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Block Height", p.L.Height)
		p.mutex.Unlock()
	case "XBFT":
		p.mutex.Lock()
		t, err := util.MakeTarget(p.Ipv4, p.Port)
		if err != nil {
			log.Println("could not make target, ", err)
		}
		s += fmt.Sprintf("\n| %s |\n", "================ PEER ================")
		s += fmt.Sprintf("|%-17s| %-20d |\n", "ID", p.ID)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Address", t)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Role", p.Role.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Round", p.ConsensusRound)
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Current Primary", p.XBFTPrimary)
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus Phase", p.XBFTPhase.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Cert[Prepare]", len(p.D.CandidateBlockCertificates[p.XBFTPrimary][plum.XBFTPhase_XBFTPrepare]))
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Threshold(P)", p.XBFTThreshold[p.XBFTPrimary][plum.XBFTPhase_XBFTPrepare])
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Cert[Commit]", len(p.D.CandidateBlockCertificates[p.XBFTPrimary][plum.XBFTPhase_XBFTCommit]))
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Threshold(C)", p.XBFTThreshold[p.XBFTPrimary][plum.XBFTPhase_XBFTCommit])
		s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus State", p.ConsensusState.String())
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Block Height", p.L.Height)
		s += fmt.Sprintf("|%-17s| %-20f |\n", "Reputation", p.ReputationBook[p.ID])
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Selected Count", p.SelectedCount)
		s += fmt.Sprintf("|%-17s| %-20d |\n", "Tentative SC", p.TentativeSelectedCount)
		p.mutex.Unlock()
	}

	return s
}

func (p peer) PrintPeer() {
	d := p.D
	switch d.ConsensusType {
	case "PBFT":
		log.Println(fmt.Sprintf("%s\n%s\n%s\n", p.String(), p.MQ.String(), d.ReservedPBFTMessage.String()))
	case "XBFT":
		log.Println(fmt.Sprintf("%s\n%s\n%s\n%s\n", p.String(), p.XBFTMQ.String(), d.ReservedXBFTMessage.String(), p.ReputationBookString()))
		util.PrintCommitteeMembers(p.D.committeeMembers)
	}
}

func (p peer) ReputationBookString() string {
	var s string
	s += fmt.Sprintf("Rep Book ===========\n")
	s += fmt.Sprintf("ID | Reputation\n")

	repByPeerID := make([]float64, len(p.ReputationBook))
	for k, v := range p.ReputationBook {
		repByPeerID[k] = v
	}

	for i, v := range repByPeerID {
		s += fmt.Sprintf("%7d | %v\n", i, v)
	}
	return s
}

func GetInstance() *peer {
	once.Do(func() {
		instance = new(peer)
	})
	return instance
}

//SendAll sends a consensus message to all peers in the address book with waiting mechanism
func SendAll(request interface{}) {
	var wg sync.WaitGroup
	peerInstance := GetInstance()

	switch cr := request.(type) {
	case *plum.PBFTRequest:
		ch := make(chan *plum.PBFTResponse, len(peerInstance.AddressBook))
		for _, address := range peerInstance.AddressBook {
			wg.Add(1)
			go sendPBFTMessage(address, cr, &wg, ch)
		}
		wg.Wait()
		close(ch)
	case *plum.XBFTRequest:
		ch := make(chan *plum.XBFTResponse, len(peerInstance.AddressBook))
		for _, address := range peerInstance.AddressBook {
			wg.Add(1)
			go sendXBFTMessage(address, cr, &wg, ch)
		}
		wg.Wait()
		close(ch)
	}
}

func SendAllExceptThisPeer(request interface{}) {
	var wg sync.WaitGroup
	peerInstance := GetInstance()

	switch cr := request.(type) {
	case *plum.PBFTRequest:
		ch := make(chan *plum.PBFTResponse, len(peerInstance.AddressBook))
		for _, address := range peerInstance.AddressBook {
			if address.PeerId == peerInstance.ID {
				continue
			}
			wg.Add(1)
			go sendPBFTMessage(address, cr, &wg, ch)
		}
		wg.Wait()
		close(ch)
	case *plum.XBFTRequest:
		ch := make(chan *plum.XBFTResponse, len(peerInstance.AddressBook))
		for _, address := range peerInstance.AddressBook {
			if address.PeerId == peerInstance.ID {
				continue
			}
			wg.Add(1)
			go sendXBFTMessage(address, cr, &wg, ch)
		}
		wg.Wait()
		close(ch)
	}
}

func SendCommitteeMembers(request interface{}) {
	var wg sync.WaitGroup
	peerInstance := GetInstance()

	switch cr := request.(type) {
	case *plum.XBFTRequest:
		ch := make(chan *plum.XBFTResponse, len(peerInstance.AddressBook))
		for _, k := range peerInstance.D.committeeMembers {
			wg.Add(1)
			go sendXBFTMessage(peerInstance.AddressBook[k.GetPeerId()], cr, &wg, ch)
		}
		wg.Wait()
		close(ch)
	}
}

//Send sends a consensus message to one specific peer
func Send(m *Connection, request interface{}) {
	switch cr := request.(type) {
	case *plum.PBFTRequest:
		c := m.consensusClient
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*15000)
		defer cancel()

		_, err := c.ServePBFTPhase(ctx, cr)

		if err != nil {
			log.Printf("could not serve: %v", err)
		}
	case *plum.XBFTRequest:
		c := m.consensusClient
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*15000)
		defer cancel()

		_, err := c.ServeXBFTPhase(ctx, cr)

		if err != nil {
			log.Printf("could not serve: %v", err)
		}
	}
}

func sendPBFTMessage(m *Connection, cr *plum.PBFTRequest, wg *sync.WaitGroup, ch chan *plum.PBFTResponse) {
	defer wg.Done()

	c := m.consensusClient

	timer := time.NewTimer(time.Second * 3)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*15000)
	defer cancel()

	r, err := c.ServePBFTPhase(ctx, cr)
	if err != nil {
		return
	}
	select {
	case <-timer.C:
		log.Println("drop this message due to sending this again for 3s")
		return
	default:
		ch <- r
	}
}

func sendXBFTMessage(m *Connection, cr *plum.XBFTRequest, wg *sync.WaitGroup, ch chan *plum.XBFTResponse) {
	defer wg.Done()

	c := m.consensusClient

	timer := time.NewTimer(time.Second * 20)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*15000)
	defer cancel()

	r, err := c.ServeXBFTPhase(ctx, cr)
	if err != nil {
		return
	}
	select {
	case <-timer.C:
		log.Println("drop this message due to sending this again for 20s", util.MakeString(cr))
		return
	default:
		ch <- r
	}
}

func (p *peer) connectAll() {
	for _, addr := range p.AddressBook {
		var target string
		var err error

		if addr.Ipv4 == "localhost" {
			//local test mode
			target, err = util.MakeTarget(addr.Ipv4, addr.Port)
		} else if addr.Ipv4 == "" {
			//local docker test mode
			target, err = util.MakeTarget(addr.ContainerName, addr.Port)
		} else {
			//distributed docker mode
			target, err = util.MakeTarget(addr.Ipv4, addr.Port)
			if addr.Ipv4 == p.AddressBook[p.ID].Ipv4 {
				//if the target is placed in the same host with the peer, use container name because they are in the same docker network
				target, err = util.MakeTarget(addr.ContainerName, addr.Port)
			}
		}

		if err != nil {
			log.Fatalf("could not make target: %v", err)
		}

		conn, err := grpc.Dial(target, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("failed to connect: %v", err)
		}
		addr.clientConn = conn
		addr.consensusClient = plum.NewConsensusClient(conn)
		addr.peerClient = plum.NewPeerClient(conn)
	}
}

func (p *peer) NewPrimary(nr uint64) uint32 {
	np := len(p.AddressBook)
	return uint32(nr % uint64(np))
}

// RetrieveTxs method should be updated to get transactions from the mempool
// Temporarily, it generates 2000 transactions with no meaning
func (p *peer) RetrieveTxs() [][]byte {
	var txs [][]byte
	for i := 0; i < 2000; i++ {
		txs = append(txs, []byte(fmt.Sprintf("tx%d", rand.Intn(100000))))
	}

	return txs
}

func (p *peer) NewCandidateBlock() *plum.Block {
	txs := p.RetrieveTxs()
	phd := block.Digest(p.L.CurrentBlockHeader())
	b := block.NewBlock(txs, phd, p.L.Height+1)
	return b
}

func (p *peer) CreateSignature(m proto.Message) []byte {
	md, err := proto.Marshal(m)
	if err != nil {
		log.Printf("could not marshal the message: %v\n", err)
	}
	sig := ed25519.Sign(p.PrivateKey, md)

	return sig
}

func (p *peer) VerifyConsensusMessageSignature(message interface{}) bool {
	switch m := message.(type) {
	case *plum.PBFTRequest:
		md, err := proto.Marshal(m.Message)
		if err != nil {
			log.Printf("could not marshal the message: %v\n", err)
		}

		if ed25519.Verify(p.AddressBook[m.Message.GetPeerId()].PublicKey, md, m.GetSignature()) {
			return true
		}

		log.Println("invalid signature for message", util.MakeString(m))
		return false
	case *plum.XBFTRequest:
		md, err := proto.Marshal(m.Message)
		if err != nil {
			log.Printf("could not marshal the message: %v\n", err)
		}

		if ed25519.Verify(p.AddressBook[m.Message.GetPeerId()].PublicKey, md, m.GetSignature()) {
			return true
		}

		//log.Println("invalid signature for message", util.MakeString(m))
		msg := m.GetMessage()
		log.Println("tried to verify with the pub key of peer", m.Message.GetPeerId(), " but failed: ", hex.EncodeToString(p.AddressBook[m.Message.GetPeerId()].PublicKey))
		log.Println("invalid signature for message", msg.GetPeerId(), " | ", msg.Phase)
		return false
	default:
		log.Panic("invalid type of message on verifying signature")
		return false
	}
}

func (p *peer) verifyCertificate(c *plum.Certificate) bool {
	if c.GetCert() == nil {
		return false
	}

	// Check phase and primary consistency
	ph, pr := c.GetCert()[0].GetMessage().GetPhase(), c.GetCert()[0].GetMessage().GetPrimaryId()
	for _, v := range c.GetCert() {
		if ph != v.GetMessage().GetPhase() || pr != v.GetMessage().GetPrimaryId() {
			return false
		}

		// Check signature
		if !p.VerifyConsensusMessageSignature(v) {
			log.Println("invalid signature detected during the verification of certificate. occurred on peer", v.GetMessage().GetPeerId())
			return false
		}

		// Check block digest from message and candidate block's
		if !block.CompareBlockDigest(v.GetMessage().GetDigest(), p.D.CandidateBlockDigests[pr]) {
			log.Println("different block digest detected during the verification of certificate. occurred on peer", v.GetMessage().GetPeerId())
			return false
		}
	}

	// Check on threshold and candidate block
	switch ph {
	case plum.XBFTPhase_XBFTRoundChange:
		log.Println("verifying round change certificate is not implemented yet")
		return true
	case plum.XBFTPhase_XBFTPrepare:
		fallthrough
	case plum.XBFTPhase_XBFTCommit:
		if len(c.GetCert()) > p.XBFTThreshold[pr][ph] {
			return true
		}
		return false
	default:
		return false
	}
}

func (p *peer) SetTimer(ph interface{}) {
	p.K.Set(p.ConsensusRound, ph)
}

func (p *peer) setXBFTPrimary(id uint32) {
	p.XBFTPrimary = id
}

func (p *peer) nextRoundCandidateBlock(rcc *plum.Certificate) *plum.Block {
	// Priority 1: has PC, CC - round change occurred at selection
	// Priority 2: has PC, but no CC - round change occurred at commit
	// Priority 3: no PC, CC - round change occurred at prepare or earlier
	// in the same priority, selection value is the casting voter

	if rcc.GetCert()[0] == nil {
		log.Panic("empty round change certificate")
	}

	// Check if is has 51%

	// Search for the round change message with the highest priority
	highestPriority := rcc.GetCert()[0]
	_, selectionValueOfHighestPriority := findXBFTPrimary(highestPriority.GetBlock().GetCommitteeMembers())

	for _, rc := range rcc.GetCert() {
		rcm := rc.GetMessage()
		highestPriorityMessage := highestPriority.GetMessage()
		if highestPriorityMessage.PreparedCertificate.GetCert() != nil && highestPriorityMessage.CommittedCertificate.GetCert() != nil {
			// compare their primary's sv of the two
			_, selectionValueOfRC := findXBFTPrimary(rc.GetBlock().GetCommitteeMembers())
			if selectionValueOfRC > selectionValueOfHighestPriority {
				highestPriority = rc
				continue
			}
		} else if highestPriorityMessage.PreparedCertificate.GetCert() != nil && highestPriorityMessage.CommittedCertificate.GetCert() == nil {
			// if rc has both certificate -> change the highest
			if rcm.GetPreparedCertificate().GetCert() != nil && rcm.GetCommittedCertificate().GetCert() != nil {
				highestPriority = rc
				continue
			}
			// if rc has prepared certificate only -> compare their primary's sv of the two
			if rcm.GetPreparedCertificate().GetCert() != nil && rcm.GetCommittedCertificate().GetCert() == nil {
				// compare their primary's sv of the two
				_, selectionValueOfRC := findXBFTPrimary(rc.GetBlock().GetCommitteeMembers())
				if selectionValueOfRC > selectionValueOfHighestPriority {
					highestPriority = rc
				}
				continue
			}
		} else if highestPriorityMessage.PreparedCertificate.GetCert() == nil && highestPriorityMessage.CommittedCertificate.GetCert() == nil {
			if rcm.GetPreparedCertificate().GetCert() != nil && rcm.GetCommittedCertificate().GetCert() != nil {
				highestPriority = rc
			} else if rcm.GetPreparedCertificate().GetCert() != nil && rcm.GetCommittedCertificate().GetCert() == nil {
				highestPriority = rc
			}
		}
	}

	// Set a new candidate block if all the round change has no certificate for both prepared and committed
	if highestPriority.GetMessage().PreparedCertificate.GetCert() == nil && highestPriority.GetMessage().CommittedCertificate.GetCert() == nil {
		roundChangedCommitteeMember := highestPriority.GetBlock().GetCommitteeMembers()
		highestPriority.Block = p.NewCandidateBlock()
		highestPriority.Block.CommitteeMembers = roundChangedCommitteeMember
	}

	return highestPriority.GetBlock()
}
