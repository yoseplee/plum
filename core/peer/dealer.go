package peer

import (
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/peer/heap"
	"github.com/yoseplee/plum/core/peer/mq"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/vrf"
	"log"
)

const (
	scheduleMq   scheduleCriteria = 0
	scheduleHeap scheduleCriteria = 1
)

type scheduleCriteria uint

type Dealer struct {
	ConsensusType               string
	MQ                          *mq.Queue
	XBFTMQ                      *mq.XQueue
	ReservedPBFTMessage         *heap.MinPBFTHeap
	ReservedXBFTMessage         *heap.MinXBFTHeap
	ReservedPrepareMessage      map[uint32][]*plum.XBFTRequest                    //for XBFT
	ReservedCommitMessage       map[uint32][]*plum.XBFTRequest                    //for XBFT
	CandidateBlock              *plum.Block                                       //for PBFT
	CandidateBlockDigest        []byte                                            //for PBFT
	CandidateBlocks             map[uint32]*plum.Block                            //for XBFT
	CandidateBlockDigests       map[uint32][]byte                                 //for XBFT
	CandidateBlockCertificates  map[uint32]map[plum.XBFTPhase][]*plum.XBFTRequest //for XBFT
	CandidateCommitteeMembers   map[uint32][]*plum.CommitteeMembers               //for XBFT
	receivedReputationSum       map[uint32]float64
	roundChangeReputationSum    float64
	roundChangeCommitteeMembers []*plum.CommitteeMembers
	roundChangeCertificate      []*plum.XBFTRequest
	committeeMembers            []*plum.CommitteeMembers
	totalReputationAtRound      float64
	stopSig                     chan struct{}
}

func NewDealer(consensusType string) *Dealer {
	p := GetInstance()
	d := &Dealer{
		ConsensusType:              consensusType,
		MQ:                         p.MQ,
		XBFTMQ:                     p.XBFTMQ,
		ReservedPBFTMessage:        heap.NewMinPBFTHeap(),
		ReservedXBFTMessage:        heap.NewMinXBFTHeap(),
		ReservedPrepareMessage:     make(map[uint32][]*plum.XBFTRequest),
		ReservedCommitMessage:      make(map[uint32][]*plum.XBFTRequest),
		CandidateBlocks:            make(map[uint32]*plum.Block),
		CandidateBlockDigests:      make(map[uint32][]byte),
		CandidateBlockCertificates: make(map[uint32]map[plum.XBFTPhase][]*plum.XBFTRequest),
		CandidateCommitteeMembers:  make(map[uint32][]*plum.CommitteeMembers),
		receivedReputationSum:      make(map[uint32]float64),
		stopSig:                    make(chan struct{}),
	}
	return d
}

func (d *Dealer) Run() {
	switch d.ConsensusType {
	case "PBFT":
		d.startPBFT()
	case "XBFT":
		d.totalReputationAtRound = GetInstance().RepSum()
		d.startXBFT()
	default:
		log.Fatalf("invalid type of consensus")
	}
}

func (d *Dealer) committeeMemberAtPrePrepare(peerID uint32) bool {
	for _, k := range d.committeeMembers {
		if k.GetPeerId() == peerID {
			return true
		}
	}
	return false
}

func (d *Dealer) setCandidateBlock(cb *plum.Block) {
	d.CandidateBlock = cb
	d.CandidateBlockDigest = block.Digest(d.CandidateBlock.Header)
}

func (d *Dealer) setXBFTCandidateBlock(peerID uint32, cb *plum.Block) {
	cbd := block.Digest(cb.Header)
	d.CandidateBlocks[peerID] = cb
	d.CandidateBlockDigests[peerID] = cbd
}

func (d *Dealer) roundCheck(message interface{}) bool {
	p := GetInstance()
	switch m := message.(type) {
	case *plum.PBFTRequest:
		//round check
		if p.ConsensusRound < m.Message.GetRound() {
			d.ReservedPBFTMessage.Push(m)
			return false
		} else if p.ConsensusRound > m.Message.GetRound() {
			//discard
			return false
			//} else if p.ConsensusRound == m.Message.GetRound() && p.PBFTPhase >= m.Message.GetPhase() && m.Message.GetPhase() != plum.ConsensusPhase_NEW_ROUND {
		} else if p.ConsensusRound == m.Message.GetRound() {

			if m.Message.GetPhase() == plum.PBFTPhase_PBFTNewRound {
				return true
			}

			if m.Message.GetPhase() == plum.PBFTPhase_PBFTRoundChange {
				return true
			}

			if p.PBFTPhase >= m.Message.GetPhase() {
				//discard
				return false
			}
		}
		return true
	case *plum.XBFTRequest:

		// check for round change messages
		if m.GetMessage().GetPhase() == plum.XBFTPhase_XBFTRoundChange && m.GetMessage().GetHeight() == p.L.Height { // check my state is round change?
			return true
		}

		// fixing round change logic
		//if m.GetMessage().GetPhase() == plum.XBFTPhase_XBFTPrePrepare && m.GetMessage().GetHeight() == p.L.Height && m.GetMessage().GetRoundChangeCertificate() != nil {
		//	log.Println("update consensus round according to the round change message with certificate from", p.ConsensusRound, " to", m.GetMessage().GetRound())
		//	p.ConsensusRound = m.GetMessage().GetRound()
		//	//log.Println("immediately start pre prepare message")
		//	//handleXBFTPrePrepare(m)
		//	return true
		//}

		//round check
		if p.ConsensusRound < m.Message.GetRound() {
			d.ReservedXBFTMessage.Push(m)
			return false
		} else if p.ConsensusRound > m.Message.GetRound() {
			//discard
			return false
			//} else if p.ConsensusRound == m.Message.GetRound() && p.PBFTPhase >= m.Message.GetPhase() && m.Message.GetPhase() != plum.ConsensusPhase_NEW_ROUND {
		} else if p.ConsensusRound == m.Message.GetRound() {
			return true
		}
		return true
	default:
		log.Panic("invalid message type on round check")
		return false
	}
}

//triggerXBFTRoundChange() is called when tie node is timed out by keeper instance
//it resets state of the peer, increase round then multicasts round change message
func (d *Dealer) triggerXBFTRoundChange() {
	p := GetInstance()
	//discard request messages at current round in the heap
	d.discardAllTheRemainedMessagesAtTheRound()

	log.Println("round change triggered: ", p.ConsensusRound, "->", p.ConsensusRound+1)
	p.ConsensusRound++
	p.XBFTPhase = plum.XBFTPhase_XBFTRoundChange

	var pi, vrfHash []byte
	var proveErr error
	if d.CandidateBlocks[p.XBFTPrimary].GetRoundChangedCommitteeMembers() == nil {
		pi, vrfHash, proveErr = vrf.Prove(p.PublicKey, p.PrivateKey, p.L.CurrentBlockHeader().MerkleRoot)
		if proveErr != nil {
			log.Println("could not prove:", proveErr)
			return
		}
	} else {
		//the block has already failed -> 2nd tie break rule should be used
		pi, vrfHash, proveErr = vrf.Prove(p.PublicKey, p.PrivateKey, p.L.CurrentBlockHeader().PrevBlockHash)
		if proveErr != nil {
			log.Println("could not prove:", proveErr)
			return
		}
	}

	selectionValue := p.selectionValue(vrfHash, p.ID)

	if Selection(selectionValue) {
		p.TentativeSelectedCount++
	}

	//get prepared certificate
	var preparedCertificate *plum.Certificate
	pCert, pCertExists := pCert(p.XBFTPrimary)
	if pCertExists {
		preparedCertificate = &plum.Certificate{Cert: pCert}
	} else {
		preparedCertificate = nil
	}

	//get committed certificate
	var committedCertificate *plum.Certificate
	cCert, cCertExists := cCert(p.XBFTPrimary)
	if cCertExists {
		committedCertificate = &plum.Certificate{Cert: cCert}
	} else {
		committedCertificate = nil
	}

	consensusMessage := &plum.XBFTMessage{
		Phase:                plum.XBFTPhase_XBFTRoundChange,
		Round:                p.ConsensusRound,
		Height:               p.L.Height,
		PeerId:               p.ID,
		Proof:                pi,
		PrimaryId:            p.XBFTPrimary,
		Digest:               p.D.CandidateBlockDigests[p.XBFTPrimary],
		PreparedCertificate:  preparedCertificate,
		CommittedCertificate: committedCertificate,
		SelectionValue:       selectionValue,
	}
	signature := p.CreateSignature(consensusMessage)
	roundChangeMessage := &plum.XBFTRequest{
		Message:   consensusMessage,
		Signature: signature,
		Block:     p.D.CandidateBlocks[p.XBFTPrimary],
	}
	p.D.handleXBFT(roundChangeMessage)
	go SendAllExceptThisPeer(roundChangeMessage)
}

func (d *Dealer) discardAllTheRemainedMessagesAtTheRound() {
	p := GetInstance()
	switch p.D.ConsensusType {
	case "PBFT":
		for {
			t, hErr := p.D.ReservedPBFTMessage.Peek()
			if hErr != nil {
				//empty heap
				break
			}

			if t.GetMessage().GetRound() == p.ConsensusRound {
				_, err := p.D.ReservedPBFTMessage.Pop()
				if err != nil {
					log.Printf("could not discard the all remained messages at round %d %v", p.ConsensusRound, err)
				}
			} else {
				break
			}
		}
	case "XBFT":
		for {
			t, hErr := p.D.ReservedXBFTMessage.Peek()
			if hErr != nil {
				//empty heap
				break
			}

			if t.GetMessage().GetRound() == p.ConsensusRound {
				_, err := p.D.ReservedXBFTMessage.Pop()
				if err != nil {
					log.Printf("could not discard the all remained messages at round %d %v", p.ConsensusRound, err)
				}
			} else {
				break
			}
		}
	}
}

func vote(phase interface{}) {
	switch ph := phase.(type) {
	case plum.PBFTPhase:
		GetInstance().PBFTVote[ph]++
	}
}

func findCommitteeMemberById(id uint32, cms []*plum.CommitteeMembers) (*plum.CommitteeMembers, error) {
	for _, cm := range cms {
		if cm.PeerId == id {
			return cm, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("there is no matching committee member with id %d", id))
}

func (p *peer) assignRoleByCommitteeMember() {
	_, involved := findCommitteeMemberById(p.ID, p.D.committeeMembers)
	if involved == nil {
		p.Role = plum.ConsensusRole_CommitteeMember
	} else {
		p.Role = plum.ConsensusRole_Backup
	}
}

func (d *Dealer) emptyCandidateBlocks() bool {
	if len(d.CandidateBlocks) == 0 {
		return true
	}
	return false
}

func findXBFTPrimary(committeeMembers []*plum.CommitteeMembers) (uint32, float64) {
	var highestValue float64
	var highestKey uint32
	for _, v := range committeeMembers {
		if v.GetSelectionValue() > highestValue {
			highestValue = v.GetSelectionValue()
			highestKey = v.GetPeerId()
		}
	}
	return highestKey, highestValue
}

func makeCert(primaryID uint32, ph plum.XBFTPhase, m *plum.XBFTRequest) {
	p := GetInstance()
	if p.D.CandidateBlockCertificates[primaryID] == nil {
		p.D.CandidateBlockCertificates[primaryID] = make(map[plum.XBFTPhase][]*plum.XBFTRequest)
	}
	p.D.CandidateBlockCertificates[primaryID][ph] = append(p.D.CandidateBlockCertificates[primaryID][ph], m)
}

func pCert(peerID uint32) ([]*plum.XBFTRequest, bool) {
	p := GetInstance()
	pCert := p.D.CandidateBlockCertificates[peerID][plum.XBFTPhase_XBFTPrepare]
	if len(pCert) > p.XBFTThreshold[peerID][plum.XBFTPhase_XBFTPrepare] {
		return pCert, true
	}
	return nil, false
}

func cCert(peerID uint32) ([]*plum.XBFTRequest, bool) {
	p := GetInstance()
	cCert := p.D.CandidateBlockCertificates[peerID][plum.XBFTPhase_XBFTCommit]
	if len(cCert) > p.XBFTThreshold[peerID][plum.XBFTPhase_XBFTCommit] {
		return cCert, true
	}
	return nil, false
}

func handleAllTheReservedPrepareMessages(p *peer) {
	rpms := p.D.ReservedPrepareMessage[p.XBFTPrimary]
	if len(rpms) == 0 {
		// no prepare messages
		return
	}
	for _, rpm := range rpms {
		p.D.handleXBFT(rpm)
	}
}

func handleAllTheReservedCommitMessages(p *peer, primary uint32) {
	rcms := p.D.ReservedCommitMessage[primary]
	if len(rcms) == 0 {
		// no commit messages
		return
	}
	for _, rcm := range rcms {
		p.D.handleXBFT(rcm)
	}
}

func verifySelect(peerID uint32, proof []byte, seed []byte, selectionValue float64) bool {
	p := GetInstance()
	verified, err := vrf.Verify(p.AddressBook[peerID].PublicKey, proof, seed)
	if err != nil {
		log.Println("invalid proof", err)
		return false
	}
	return verified
}
