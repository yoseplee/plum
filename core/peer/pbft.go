package peer

import (
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/plum"
	"log"
)

func (d *Dealer) startPBFT() {
	log.Printf("Dealer has started up on peer %d | type: %s\n", GetInstance().ID, d.ConsensusType)
	for {
		select {
		case <-d.stopSig:
			log.Printf("Dealer stopped on peer %d\n", GetInstance().ID)
			return
		default:
			d.schedulePBFT(scheduleMq)
		}
		d.schedulePBFT(scheduleHeap)
	}
}

func (d *Dealer) schedulePBFT(criteria scheduleCriteria) {
	switch criteria {
	case scheduleMq:
		m, err := d.MQ.Pop()
		if err != nil {
			break
		}
		d.handlePBFT(m.D)
	case scheduleHeap:
		m, err := d.ReservedPBFTMessage.Pop()
		if err != nil {
			break
		}
		d.handlePBFT(m)
	}
}

func (d *Dealer) handlePBFT(m *plum.PBFTRequest) {
	p := GetInstance()
	//log.Printf("[p %d] handle %s | current_round(%d), current_phase(%s), n_mq(%d), n_reserved(%d)\n", p.ID, util.MakeString(m), p.ConsensusRound, p.PBFTPhase.String(), d.MQ.GetN(), d.Reserved.Last)

	if !d.roundCheck(m) || !p.VerifyConsensusMessageSignature(m) {
		return
	}

	switch m.Message.GetPhase() {
	case plum.PBFTPhase_PBFTNewRound:
		handlePBFTNewRound(m)
	case plum.PBFTPhase_PBFTPrePrepare:
		handlePBFTPrePrepare(m)
	case plum.PBFTPhase_PBFTPrepare:
		handlePBFTPrepare(m)
	case plum.PBFTPhase_PBFTCommit:
		handlePBFTCommit(m)
	case plum.PBFTPhase_PBFTRoundChange:
		handlePBFTRoundChange(m)
	default:
		log.Println("failed to handle consensus message due to the invalid phase")
	}
}

func handlePBFTNewRound(m *plum.PBFTRequest) {
	p := GetInstance()
	if p.PBFTPhase == plum.PBFTPhase_PBFTNewRound {
		//assign role
		p.Role = plum.ConsensusRole_Primary

		p.D.setCandidateBlock(GetInstance().NewCandidateBlock())

		//send all
		consensusMessage := &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTPrePrepare,
			Round:  p.ConsensusRound,
			Digest: p.D.CandidateBlockDigest,
			PeerId: p.ID,
		}
		signature := p.CreateSignature(consensusMessage)
		go SendAll(&plum.PBFTRequest{
			Message:   consensusMessage,
			Signature: signature,
			Block:     p.D.CandidateBlock,
		})
	} else {
		p.D.ReservedPBFTMessage.Push(m)
	}
}

func handlePBFTPrePrepare(m *plum.PBFTRequest) {
	p := GetInstance()
	if p.PBFTPhase != plum.PBFTPhase_PBFTNewRound {
		//reserve
		p.D.ReservedPBFTMessage.Push(m)
		return
	}

	p.PBFTPhase = plum.PBFTPhase_PBFTPrePrepare
	p.SetTimer(m.Message.Phase)

	if p.Role == plum.ConsensusRole_Primary {
		return
	}

	//set received block as candidate block: backups only
	p.D.setCandidateBlock(m.Block)

	//verify digest - may redundant
	if !block.CompareBlockDigest(p.D.CandidateBlockDigest, m.GetMessage().GetDigest()) {
		return
	}

	//send all
	consensusMessage := &plum.PBFTMessage{
		Phase:  plum.PBFTPhase_PBFTPrepare,
		Round:  p.ConsensusRound,
		Digest: p.D.CandidateBlockDigest,
		PeerId: p.ID,
	}
	signature := p.CreateSignature(consensusMessage)
	go SendAll(&plum.PBFTRequest{
		Message:   consensusMessage,
		Signature: signature,
	})
}

func handlePBFTPrepare(m *plum.PBFTRequest) {
	p := GetInstance()
	if p.PBFTPhase != plum.PBFTPhase_PBFTPrePrepare {
		p.D.ReservedPBFTMessage.Push(m)
		return
	}

	//verify
	if !block.CompareBlockDigest(m.GetMessage().GetDigest(), p.D.CandidateBlockDigest) {
		return
	}

	if p.PBFTVote[plum.PBFTPhase_PBFTPrepare] == 0 {
		//start the timer because this is the first time to prepare
		p.SetTimer(m.Message.Phase)
	}

	vote(plum.PBFTPhase_PBFTPrepare)

	if p.PBFTVote[plum.PBFTPhase_PBFTPrepare] > p.PBFTThreshold[plum.PBFTPhase_PBFTPrepare] {
		p.PBFTPhase = plum.PBFTPhase_PBFTPrepare

		consensusMessage := &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTCommit,
			Round:  p.ConsensusRound,
			Digest: p.D.CandidateBlockDigest,
			PeerId: p.ID,
		}
		signature := p.CreateSignature(consensusMessage)
		go SendAll(&plum.PBFTRequest{
			Message:   consensusMessage,
			Signature: signature,
		})
	}
}

func handlePBFTCommit(m *plum.PBFTRequest) {
	p := GetInstance()
	if p.PBFTPhase != plum.PBFTPhase_PBFTPrepare {
		p.D.ReservedPBFTMessage.Push(m)
		return
	}

	if p.PBFTVote[plum.PBFTPhase_PBFTCommit] == 0 {
		//start the timer because this is the first time to prepare
		p.SetTimer(m.Message.Phase)
	}

	//verify
	if !block.CompareBlockDigest(m.GetMessage().GetDigest(), p.D.CandidateBlockDigest) {
		return
	}

	vote(plum.PBFTPhase_PBFTCommit)

	if p.PBFTVote[plum.PBFTPhase_PBFTCommit] > p.PBFTThreshold[plum.PBFTPhase_PBFTCommit] {

		//append block to the ledger
		appendErr := p.L.Append(p.D.CandidateBlock)
		if appendErr != nil {
			log.Printf("could not append: %v\n", appendErr)
			return
		}

		//update and reset attributes in peer
		p.ConsensusRound++
		p.PBFTVote = make(map[plum.PBFTPhase]int)
		p.PBFTPhase = plum.PBFTPhase_PBFTNewRound
		p.D.CandidateBlock = nil
		p.D.CandidateBlockDigest = nil

		//set timer
		p.SetTimer(plum.PBFTPhase_PBFTNewRound)

		if p.Role == plum.ConsensusRole_Primary {
			consensusMessage := &plum.PBFTMessage{
				Phase:  plum.PBFTPhase_PBFTNewRound,
				Round:  p.ConsensusRound,
				PeerId: p.ID,
			}
			signature := p.CreateSignature(consensusMessage)
			go Send(p.AddressBook[p.ID], &plum.PBFTRequest{
				Message:   consensusMessage,
				Signature: signature,
			})
		}
	}
}

func handlePBFTRoundChange(m *plum.PBFTRequest) {
	p := GetInstance()
	//vote & count
	vote(plum.PBFTPhase_PBFTRoundChange)
	//if 2f+1 && this peer is the new peer of the next round -> send pre-prepare message to all
	if p.PBFTVote[plum.PBFTPhase_PBFTRoundChange] > p.PBFTThreshold[plum.PBFTPhase_PBFTRoundChange] {
		p.PBFTVote = make(map[plum.PBFTPhase]int)
		p.PBFTPhase = plum.PBFTPhase_PBFTNewRound
		p.SetTimer(plum.PBFTPhase_PBFTNewRound)

		//if this peer is the new primary for the next round, start the next round
		if p.ID == p.NewPrimary(m.Message.GetRound()) {
			consensusMessage := &plum.PBFTMessage{
				Phase:  plum.PBFTPhase_PBFTNewRound,
				Round:  p.ConsensusRound,
				PeerId: p.ID,
			}
			signature := p.CreateSignature(consensusMessage)
			go Send(p.AddressBook[p.ID], &plum.PBFTRequest{
				Message:   consensusMessage,
				Signature: signature,
			})
		}
	}
}

//triggerPBFTRoundChange() is called when tie node is timed out by keeper instance
//it resets state of the peer, increase round then multicasts round change message
func (d *Dealer) triggerPBFTRoundChange() {
	p := GetInstance()
	//discard request messages at current round in the heap
	d.discardAllTheRemainedMessagesAtTheRound()

	//peer update to set for the next round
	p.PBFTPhase = plum.PBFTPhase_PBFTNewRound
	p.ConsensusRound++
	p.Primary = p.NewPrimary(p.ConsensusRound)
	if p.ID == p.Primary {
		p.Role = plum.ConsensusRole_Primary
	} else {
		p.Role = plum.ConsensusRole_Backup
	}

	//send all: ROUND CHANGE
	consensusMessage := &plum.PBFTMessage{
		Phase:  plum.PBFTPhase_PBFTRoundChange,
		Round:  p.ConsensusRound,
		PeerId: p.ID,
	}
	signature := p.CreateSignature(consensusMessage)
	go SendAll(&plum.PBFTRequest{
		Message:   consensusMessage,
		Signature: signature,
	})
}
