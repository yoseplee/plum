package peer

import (
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"github.com/yoseplee/vrf"
	"log"
)

func (d *Dealer) startXBFT() {
	log.Printf("Dealer has started up on peer %d | type: %s\n", GetInstance().ID, d.ConsensusType)
	for {
		select {
		case <-d.stopSig:
			log.Printf("Dealer stopped on peer %d\n", GetInstance().ID)
			return
		default:
			d.scheduleXBFT(scheduleMq)
		}
		d.scheduleXBFT(scheduleHeap)
	}
}

func (d *Dealer) scheduleXBFT(criteria scheduleCriteria) {
	switch criteria {
	case scheduleMq:
		m, err := d.XBFTMQ.Pop()
		if err != nil {
			break
		}
		d.handleXBFT(m.D)
	case scheduleHeap:
		m, err := d.ReservedXBFTMessage.Pop()
		if err != nil {
			break
		}
		d.handleXBFT(m)
	}
}

func (d *Dealer) handleXBFT(m *plum.XBFTRequest) {
	p := GetInstance()
	//log.Printf("[p %d] handle %s | current_round(%d), current_phase(%s), n_mq(%d), n_reserved(%d)\n", p.ID, util.MakeString(m), p.ConsensusRound, p.XBFTPhase.String(), d.XBFTMQ.GetN(), d.ReservedXBFTMessage.Last)
	//p.PrintPeer()

	if !d.roundCheck(m) || !p.VerifyConsensusMessageSignature(m) {
		return
	}

	switch m.Message.GetPhase() {
	case plum.XBFTPhase_XBFTPrePrepare:
		handleXBFTPrePrepare(m)
	case plum.XBFTPhase_XBFTPrepare:
		handleXBFTPrepare(m)
	case plum.XBFTPhase_XBFTCommit:
		handleXBFTCommit(m)
	case plum.XBFTPhase_XBFTSelect:
		handleXBFTSelect(m)
	case plum.XBFTPhase_XBFTRoundChange:
		handleXBFTRoundChange(m)
	default:
		log.Println("failed to handle consensus message due to the invalid phase")
	}
}

func handleXBFTRoundChange(m *plum.XBFTRequest) {
	p := GetInstance()
	senderID := m.GetMessage().GetPeerId()

	//if p.D.CandidateBlocks[p.XBFTPrimary].GetRoundChangedCommitteeMembers() == nil {
	//	if !verifySelect(
	//		m.GetMessage().GetPeerId(),
	//		m.GetMessage().GetProof(),
	//		p.L.CurrentBlockHeader().MerkleRoot,
	//		m.GetMessage().GetSelectionValue(),
	//	) {
	//		log.Println("invalid verification of selection during the round change(seed: merkle root) at round", m.GetMessage().GetRound())
	//		return
	//	}
	//} else {
	//	if !verifySelect(
	//		m.GetMessage().GetPeerId(),
	//		m.GetMessage().GetProof(),
	//		p.L.CurrentBlockHeader().PrevBlockHash,
	//		m.GetMessage().GetSelectionValue(),
	//	) {
	//		log.Println("invalid verification of selection during the round change(seed: prev block hash) at round", m.GetMessage().GetRound())
	//		return
	//	}
	//}

	//store message into the log
	p.XBFTMessageLog.Store(m)

	p.D.roundChangeReputationSum += p.ReputationBook[senderID]
	selectionValue := m.GetMessage().GetSelectionValue()
	if Selection(selectionValue) {
		p.D.roundChangeCommitteeMembers = append(p.D.roundChangeCommitteeMembers, &plum.CommitteeMembers{
			PeerId:         m.GetMessage().GetPeerId(),
			Round:          m.GetMessage().GetRound(),
			SelectionValue: m.GetMessage().GetSelectionValue(),
			Proof:          m.GetMessage().GetProof(),
		})
	}

	// forming a round change certificate... don't need distinctive primary id for data
	p.D.roundChangeCertificate = append(p.D.roundChangeCertificate, m)

	p.K.setXBFTTimer(plum.XBFTPhase_XBFTRoundChange)

	repRatio := p.D.roundChangeReputationSum / p.D.totalReputationAtRound
	//log.Println("reputation ratio for round change:", repRatio)
	//log.Println("collected number of committee member:", len(p.D.roundChangeCommitteeMembers), "/", p.minimumCommitteeSize())

	if repRatio > 0.5 && len(p.D.roundChangeCommitteeMembers) >= p.minimumCommitteeSize() {

		// 1. Increase round
		var highestRoundAtRcc uint64
		for _, rc := range p.D.roundChangeCertificate {
			rcRound := rc.GetMessage().GetRound()
			if highestRoundAtRcc < rcRound {
				highestRoundAtRcc = rcRound
			}
		}

		if p.ConsensusRound < highestRoundAtRcc {
			log.Printf("update consensus round from %d to %d because there was higher consensus round in the round change certificate", p.ConsensusRound, highestRoundAtRcc)
			p.ConsensusRound = highestRoundAtRcc
		}

		log.Printf("Handle Round change %d -> %d", p.ConsensusRound, p.ConsensusRound+1)
		p.ConsensusRound++

		// 2. Calculate Summation of Reputation at round
		p.D.totalReputationAtRound = GetInstance().RepSum()

		// 3. Set committee member from roundChangeCommitteeMembers
		p.D.committeeMembers = p.D.roundChangeCommitteeMembers

		// 4. Reset node states
		p.ConsensusState = plum.ConsensusState_PrePrepared
		p.XBFTPhase = plum.XBFTPhase_XBFTPrePrepare
		p.D.ReservedPrepareMessage = make(map[uint32][]*plum.XBFTRequest)
		p.D.ReservedCommitMessage = make(map[uint32][]*plum.XBFTRequest)
		p.D.CandidateBlocks = make(map[uint32]*plum.Block)
		p.D.CandidateBlockDigests = make(map[uint32][]byte)
		p.D.CandidateBlockCertificates = make(map[uint32]map[plum.XBFTPhase][]*plum.XBFTRequest)
		p.D.CandidateCommitteeMembers = make(map[uint32][]*plum.CommitteeMembers)
		p.D.receivedReputationSum = make(map[uint32]float64)
		p.D.totalReputationAtRound = p.RepSum()
		p.D.roundChangeReputationSum = 0.0
		p.D.roundChangeCommitteeMembers = nil

		// 5. Set node role based on the selection result
		p.assignRoleByCommitteeMember()

		// 6. Find a new primary among the collected selection results
		primary, _ := findXBFTPrimary(p.D.committeeMembers)

		// 7. Set a new primary
		p.setXBFTPrimary(primary)

		// 8. If this node is the primary
		if p.ID == primary {
			// 8.1. Change its role as the primary
			p.Role = plum.ConsensusRole_Primary

			// 8.2. Create a new Candidate Block (OR from the prepared certification)
			p.D.setXBFTCandidateBlock(p.ID, p.nextRoundCandidateBlock(&plum.Certificate{Cert: p.D.roundChangeCertificate}))

			log.Println("newly added round changed committee member")
			util.PrintCommitteeMembers(p.D.CandidateBlocks[p.ID].CommitteeMembers)

			// 8.3. Append committee member to round changed committee members
			for _, cm := range p.D.CandidateBlocks[p.ID].CommitteeMembers {
				p.D.CandidateBlocks[p.ID].RoundChangedCommitteeMembers = append(p.D.CandidateBlocks[p.ID].RoundChangedCommitteeMembers, cm)
			}

			// 8.4. Set block's committee member as the primary's committee member
			p.D.CandidateBlocks[p.ID].CommitteeMembers = p.D.committeeMembers

			// 8.5. Multicast a Pre-prepare message to all
			consensusMessage := &plum.XBFTMessage{
				Phase:                  plum.XBFTPhase_XBFTPrePrepare,
				Round:                  p.ConsensusRound,
				Height:                 p.L.Height,
				Digest:                 p.D.CandidateBlockDigests[p.ID],
				PeerId:                 p.ID,
				RoundChangeCertificate: &plum.Certificate{Cert: p.D.roundChangeCertificate},
			}
			signature := p.CreateSignature(consensusMessage)
			prePrepareMessage := &plum.XBFTRequest{
				Message:   consensusMessage,
				Signature: signature,
				Block:     p.D.CandidateBlocks[p.ID],
			}
			p.D.handleXBFT(prePrepareMessage)
			go SendAllExceptThisPeer(prePrepareMessage)
		}
	}
}

func handleXBFTPrePrepare(m *plum.XBFTRequest) {
	p := GetInstance()
	senderID := m.GetMessage().GetPeerId()
	receivedPrimary, err := findCommitteeMemberById(senderID, m.GetBlock().CommitteeMembers)
	if err != nil {
		log.Println("could not find a received primary", err)
	}

	//util.DebugMsg(fmt.Sprintf("received pre-prepare message from peer-%d with %v", m.GetMessage().GetPeerId(), receivedPrimary.SelectionValue))

	// 1. verify the candidate block - not implemented yet
	//if !block.Verify(m.GetBlock()) {
	//	log.Panic("invalid block")
	//}

	// 2. Store Message
	p.XBFTMessageLog.Store(m)

	// 3. Set Candidate Block
	p.D.setXBFTCandidateBlock(senderID, m.GetBlock())

	// 4. Set threshold for received primary's committee members
	p.setXBFTThreshold(senderID, m.GetBlock().GetCommitteeMembers())

	// 5. if the received Primary is the recognized Primary,
	if p.XBFTPrimary == senderID {
		// 5.1. Change current Committee member according to candidate block from the received primary
		p.D.committeeMembers = m.GetBlock().GetCommitteeMembers()
		//log.Println("received from the recognized primary")
		//util.PrintCommitteeMembers(p.D.committeeMembers)

		// 5.3. Change the role according to committee member in the candidate block of received primary
		p.assignRoleByCommitteeMember()

		// 5.5. Set timer
		p.SetTimer(m.Message.Phase)

		// [EXPERIMENT] verify the candidate block - if peer 0 or 1 => discard!
		if senderID == 0 || senderID == 1 {
			log.Println("got invalid block from a malicious node -", senderID)
			return
		}

		// 5.4. Set Phase: pre-prepare
		p.XBFTPhase = plum.XBFTPhase_XBFTPrePrepare
		p.ConsensusState = plum.ConsensusState_Idle

		// 5.6. Handle all the reserved prepare messages
		handleAllTheReservedPrepareMessages(p)
	} else {
		//util.DebugMsg("received pre-prepare message from not recognized node, comparing the two SVs")

		// 6. Compare current Primary's SV(Selection value) and received Primary's SV
		currentPrimary, err := findCommitteeMemberById(p.XBFTPrimary, p.D.committeeMembers)
		if err != nil {
			log.Println("could not find current primary", err)
		}
		selectionValueOfCurrentPrimary := currentPrimary.SelectionValue
		selectionValueOfReceivedPrimary := receivedPrimary.SelectionValue

		// 7. If the received Primary has higher SV than current Primary,
		if selectionValueOfReceivedPrimary > selectionValueOfCurrentPrimary {
			//log.Println("received from the primary with the higher selection value")

			// 7.1. If the peer has committed certification, keep the current primary instead of changing
			if p.ConsensusState == plum.ConsensusState_Committed {
				//log.Println("but keep current primary because this peer has already been sent select message to all")
				handleAllTheReservedCommitMessages(p, senderID)
				return
			}

			// 7.2. Change current Primary to received Primary
			p.setXBFTPrimary(m.GetMessage().GetPeerId())

			// 7.3. Change current Committee member according to candidate block from the received primary
			p.D.committeeMembers = m.GetBlock().GetCommitteeMembers()

			// 7.4. Change the role according to committee members in the candidate block of received primary
			p.assignRoleByCommitteeMember()

			// 7.5. Set timer
			p.SetTimer(m.Message.Phase)

			// [EXPERIMENT] verify the candidate block - if peer 0 or 1 => discard!
			if senderID == 0 || senderID == 1 {
				log.Println("got invalid block from a malicious node -", senderID)
				return
			}

			// 7.6. Set Phase: pre-prepare as the primary has changed
			p.XBFTPhase = plum.XBFTPhase_XBFTPrePrepare
			p.ConsensusState = plum.ConsensusState_Idle

			// 7.7. Handle all the reserved prepare messages
			handleAllTheReservedPrepareMessages(p)
		}
	}

	// 8. Send Prepare to all the committee members (Committee member only excluding the primary)
	_, exists := p.D.CandidateBlocks[p.XBFTPrimary]
	if p.Role == plum.ConsensusRole_CommitteeMember && exists && p.ID != p.XBFTPrimary {
		consensusMessage := &plum.XBFTMessage{
			Phase:     plum.XBFTPhase_XBFTPrepare,
			Round:     p.ConsensusRound,
			Height:    p.L.Height,
			Digest:    p.D.CandidateBlockDigests[p.XBFTPrimary],
			PeerId:    p.ID,
			PrimaryId: p.XBFTPrimary,
		}
		signature := p.CreateSignature(consensusMessage)
		go SendCommitteeMembers(&plum.XBFTRequest{
			Message:   consensusMessage,
			Signature: signature,
		})
	}

	// 9. Handle all the commit messages which is corresponding to the pre-prepare messages
	handleAllTheReservedCommitMessages(p, senderID)

	// 10. Reset round change certificate
	p.D.roundChangeCertificate = nil
}

func handleXBFTPrepare(m *plum.XBFTRequest) {
	p := GetInstance()
	receivedPrimary := m.GetMessage().GetPrimaryId()

	// 1. Check phase: only committee member can process prepare messages
	if p.XBFTPhase != plum.XBFTPhase_XBFTPrePrepare {
		p.D.ReservedPrepareMessage[receivedPrimary] = append(p.D.ReservedPrepareMessage[receivedPrimary], m)
		return
	}

	// 2. If the received primary is different to the current primary, reserve the message
	if p.XBFTPrimary != receivedPrimary {
		p.D.ReservedPrepareMessage[receivedPrimary] = append(p.D.ReservedPrepareMessage[receivedPrimary], m)
		return
	}

	// 3. If the received primary is equal to current primary, but didn't receive pre-prepare from the primary yet, reserve the message
	if p.D.CandidateBlockDigests[p.XBFTPrimary] == nil {
		//or if the block is nil -> reserve
		p.D.ReservedPrepareMessage[receivedPrimary] = append(p.D.ReservedPrepareMessage[receivedPrimary], m)
		return
	}

	// 4. Check block digest
	if !block.CompareBlockDigest(p.D.CandidateBlockDigests[p.XBFTPrimary], m.GetMessage().GetDigest()) {
		log.Println("different block digest detected. current primary is", p.XBFTPrimary, ", received primary is", m.GetMessage().GetPrimaryId())
		return
	}

	// 5. Store the message
	p.XBFTMessageLog.Store(m)

	// 6. Add the message to the preparedCertificate
	makeCert(p.XBFTPrimary, plum.XBFTPhase_XBFTPrepare, m)

	// 7. Set timer
	pCert, pCertExists := pCert(p.XBFTPrimary)
	if len(pCert) == 0 {
		//start the timer because this is the first time to prepare
		p.SetTimer(plum.XBFTPhase_XBFTPrepare)
	}

	// 8. If preparedCertification is formed successfully,
	if pCertExists {

		// 8.1. Change node state to be XBFT_Prepare
		p.XBFTPhase = plum.XBFTPhase_XBFTPrepare

		// 8.2. Multicast Commit message to all (including itself)
		consensusMessage := &plum.XBFTMessage{
			Phase:               plum.XBFTPhase_XBFTCommit,
			Round:               p.ConsensusRound,
			Height:              p.L.Height,
			Digest:              p.D.CandidateBlockDigests[p.XBFTPrimary],
			PeerId:              p.ID,
			PrimaryId:           p.XBFTPrimary,
			PreparedCertificate: &plum.Certificate{Cert: pCert},
		}
		signature := p.CreateSignature(consensusMessage)
		commitMessage := &plum.XBFTRequest{
			Message:   consensusMessage,
			Signature: signature,
		}
		// [EXPERIMENT] malicious node doesn't send prepare message
		if !p.Malicious {
			p.D.handleXBFT(commitMessage)
			go SendAllExceptThisPeer(commitMessage)
		}
	}
}

func handleXBFTCommit(m *plum.XBFTRequest) {
	p := GetInstance()
	receivedPrimary := m.GetMessage().GetPrimaryId()

	// 1. Check corresponding candidate block is exists
	_, candidateBlockExists := p.D.CandidateBlockDigests[receivedPrimary]
	if !candidateBlockExists {
		//reserve until the corresponding pre-prepare message arrives
		p.D.ReservedCommitMessage[receivedPrimary] = append(p.D.ReservedCommitMessage[receivedPrimary], m)
		return
	}

	// 2. Check block digest
	if !block.CompareBlockDigest(p.D.CandidateBlockDigests[receivedPrimary], m.GetMessage().GetDigest()) {
		log.Println("different block digest is detected at", m.GetMessage().GetPhase())
		return
	}

	// verify received prepared certificate
	p.verifyCertificate(m.GetMessage().GetPreparedCertificate())

	// 3. If it is from the recognized primary
	if receivedPrimary == p.XBFTPrimary {

		// 3.1. If the peer is a committee member or primary but don't have prepared certification yet, reserve this message
		pCert, pCertExists := pCert(p.XBFTPrimary)
		if (p.Role == plum.ConsensusRole_CommitteeMember || p.Role == plum.ConsensusRole_Primary) && !pCertExists && receivedPrimary == p.XBFTPrimary {
			p.D.ReservedXBFTMessage.Push(m)
			return
		}

		// 3.2. Add this message for committed certification
		makeCert(receivedPrimary, plum.XBFTPhase_XBFTCommit, m)

		// 3.3. Set timer
		cCert, cCertExists := cCert(p.XBFTPrimary)
		if len(cCert) == 0 {
			//start the timer because this is the first time to prepare
			p.SetTimer(m.Message.Phase)
		}

		// 3.4. If the peer has committed certification and have not sent select message yet,
		if cCertExists && p.ConsensusState != plum.ConsensusState_Committed {

			// 3.4.1. Change node state to be XBFT_Commit
			p.XBFTPhase = plum.XBFTPhase_XBFTCommit
			p.ConsensusState = plum.ConsensusState_Committed

			// 3.4.2. Generate vrf hash and corresponding proof
			pi, vrfHash, proveErr := vrf.Prove(p.PublicKey, p.PrivateKey, p.D.CandidateBlockDigests[p.XBFTPrimary])
			if proveErr != nil {
				log.Println("could not prove:", proveErr)
				return
			}

			// 3.4.3. Calculate Selection Value using vrf hash from the block digest
			selectionValue := p.selectionValue(vrfHash, p.ID)

			// 3.4.4. Increase Tentative Selected Count because this peer has selected as a committee member for the next round
			if Selection(selectionValue) {
				p.TentativeSelectedCount++
			}

			// 3.4.5. Multicast Selection Result
			consensusMessage := &plum.XBFTMessage{
				Phase:                plum.XBFTPhase_XBFTSelect,
				Round:                p.ConsensusRound,
				Height:               p.L.Height,
				PeerId:               p.ID,
				PrimaryId:            p.XBFTPrimary,
				Digest:               p.D.CandidateBlockDigests[p.XBFTPrimary],
				SelectionValue:       selectionValue,
				Proof:                pi,
				PreparedCertificate:  &plum.Certificate{Cert: pCert},
				CommittedCertificate: &plum.Certificate{Cert: cCert},
			}
			signature := p.CreateSignature(consensusMessage)
			selectMessage := &plum.XBFTRequest{
				Message:   consensusMessage,
				Signature: signature,
			}
			// [EXPERIMENT] malicious node doesn't send commit message
			if !p.Malicious {
				p.D.handleXBFT(selectMessage)
				go SendAllExceptThisPeer(selectMessage)
			}
		}
	} else {
		// 3.5. If commit certification is not formed yet, add the message to form the certification
		makeCert(receivedPrimary, plum.XBFTPhase_XBFTCommit, m)
	}
	// 4. Store message
	p.XBFTMessageLog.Store(m)
}

func handleXBFTSelect(m *plum.XBFTRequest) {
	p := GetInstance()
	receivedPrimaryID := m.GetMessage().GetPrimaryId()

	// 1. Check corresponding candidate block is exists
	_, exists := p.D.CandidateBlockDigests[receivedPrimaryID]
	if !exists {
		//reserve until the corresponding pre-prepare message arrives
		p.D.ReservedXBFTMessage.Push(m)
		return
	} else if !block.CompareBlockDigest(p.D.CandidateBlockDigests[receivedPrimaryID], m.GetMessage().GetDigest()) {
		// 2. Check block digest
		log.Println("different digest of block is detected at", m.GetMessage().GetPhase())
		return
	}

	// 3. If it doesn't have committed certificate, reserve
	_, cCertExists := cCert(receivedPrimaryID)
	if !cCertExists {
		//reserve
		p.D.ReservedXBFTMessage.Push(m)
		return
	}

	p.SetTimer(plum.XBFTPhase_XBFTSelect)

	// verify received prepared certificate
	p.verifyCertificate(m.GetMessage().GetPreparedCertificate())

	// verify received committed certificate
	p.verifyCertificate(m.GetMessage().GetCommittedCertificate())

	// 4. If has committed certificate, proceed
	if cCertExists {

		if !verifySelect(
			m.GetMessage().GetPeerId(),
			m.GetMessage().GetProof(),
			p.D.CandidateBlockDigests[receivedPrimaryID],
			m.GetMessage().GetSelectionValue(),
		) {
			log.Println("invalid verification of selection at round", m.GetMessage().GetRound())
			return
		}

		// 4.1. Store the message
		p.XBFTMessageLog.Store(m)

		// 4.2. Add to candidateCommitteeMember
		selectionValue := m.GetMessage().GetSelectionValue()
		if Selection(selectionValue) {
			p.D.CandidateCommitteeMembers[receivedPrimaryID] = append(p.D.CandidateCommitteeMembers[receivedPrimaryID], &plum.CommitteeMembers{
				PeerId:         m.GetMessage().GetPeerId(),
				Round:          m.GetMessage().GetRound(),
				SelectionValue: m.GetMessage().GetSelectionValue(),
				Proof:          m.GetMessage().GetProof(),
			})
		}

		// 4.3. Sum reputation of the sender
		p.D.receivedReputationSum[receivedPrimaryID] += p.ReputationBook[m.GetMessage().GetPeerId()]
		repRatio := p.D.receivedReputationSum[receivedPrimaryID] / p.D.totalReputationAtRound
		//log.Println("Rep ratio at selection for", receivedPrimaryID, ":", repRatio)
		//log.Println("collected number of committee member:", len(p.D.CandidateCommitteeMembers[receivedPrimaryID]), "/", p.minimumCommitteeSize())

		// 4.4. If it has collected messages from the all nodes on the same primary but the number of committee size is lacking, let round change
		if repRatio > 0.99999999 && len(p.D.CandidateCommitteeMembers[receivedPrimaryID]) < p.minimumCommitteeSize() {
			log.Println("committee size is lacking")
			p.D.triggerXBFTRoundChange()
			return
		}

		var reputationSumForAll float64
		for _, v := range p.D.receivedReputationSum {
			reputationSumForAll += v
		}
		repRatioForAll := reputationSumForAll / p.D.totalReputationAtRound

		// 4.5. If it has collected messages from the all nodes on different primaries but the number of committee size is still lacking, let round change
		// TODO: have to implement to search because any of them may have sufficient number of committee size
		if repRatioForAll > 0.99999999 && len(p.D.CandidateCommitteeMembers[receivedPrimaryID]) < p.minimumCommitteeSize() {
			log.Println("committee size is lacking")
			p.D.triggerXBFTRoundChange()
			return
		}

		// 4.6. If the summation is larger than 50% of total reputation
		if repRatio > 0.5 && len(p.D.CandidateCommitteeMembers[receivedPrimaryID]) >= p.minimumCommitteeSize() {

			// 4.6.1. Append the block
			log.Printf("Append Block height: %d, from: %d | at round %d\n", p.L.Height, receivedPrimaryID, p.ConsensusRound)
			err := p.L.Append(p.D.CandidateBlocks[receivedPrimaryID])
			if err != nil {
				p.PrintPeer()
				log.Fatalf("could not append block of %d: %v", receivedPrimaryID, err)
			}

			// 4.6.2. Increase Selected Count
			_, involved := findCommitteeMemberById(p.ID, p.D.CandidateBlocks[receivedPrimaryID].CommitteeMembers)
			if involved == nil {
				p.SelectedCount++
			}

			// 4.6.3. Update Reputation, Note that the committee member is not updated one in the phase
			p.RepDecrease(p.D.CandidateBlocks[receivedPrimaryID].RoundChangedCommitteeMembers)
			p.RepIncrease(p.D.CandidateBlocks[receivedPrimaryID].CommitteeMembers)

			// 4.6.4. Increase Round
			p.ConsensusRound++

			// 4.6.5. Calculate Summation of Reputation at that round
			p.D.totalReputationAtRound = GetInstance().RepSum()

			// 4.6.6. Set committeeMember from candidateCommitteeMember
			p.D.committeeMembers = p.D.CandidateCommitteeMembers[receivedPrimaryID]

			// 4.6.7. Reset node states
			p.ConsensusState = plum.ConsensusState_PrePrepared
			p.XBFTPhase = plum.XBFTPhase_XBFTPrePrepare
			p.D.ReservedPrepareMessage = make(map[uint32][]*plum.XBFTRequest)
			p.D.ReservedCommitMessage = make(map[uint32][]*plum.XBFTRequest)
			p.D.CandidateBlocks = make(map[uint32]*plum.Block)
			p.D.CandidateBlockDigests = make(map[uint32][]byte)
			p.D.CandidateBlockCertificates = make(map[uint32]map[plum.XBFTPhase][]*plum.XBFTRequest)
			p.D.CandidateCommitteeMembers = make(map[uint32][]*plum.CommitteeMembers)
			p.D.receivedReputationSum = make(map[uint32]float64)
			p.D.totalReputationAtRound = p.RepSum()
			p.D.roundChangeReputationSum = 0.0
			p.D.roundChangeCommitteeMembers = nil

			p.SetTimer(plum.XBFTPhase_XBFTPrePrepare)

			// 4.6.8. Set node role based on the selection result
			p.assignRoleByCommitteeMember()

			// 4.6.9. Find a new primary among the collected selection results
			primary, _ := findXBFTPrimary(p.D.committeeMembers)

			// 4.6.10. Set a new primary
			p.setXBFTPrimary(primary)

			// 4.6.11. If this node is the primary among the collected selection results
			if p.ID == primary {
				// 4.6.11.1. Change its role to be the primary
				p.Role = plum.ConsensusRole_Primary

				// 4.6.11.2. Create a new Candidate Block
				p.D.setXBFTCandidateBlock(p.ID, p.NewCandidateBlock())
				p.D.CandidateBlocks[p.ID].CommitteeMembers = p.D.committeeMembers

				// 4.6.11.3. Multicast a Pre-prepare message to all
				consensusMessage := &plum.XBFTMessage{
					Phase:  plum.XBFTPhase_XBFTPrePrepare,
					Round:  p.ConsensusRound,
					Height: p.L.Height,
					Digest: p.D.CandidateBlockDigests[p.ID],
					PeerId: p.ID,
				}
				signature := p.CreateSignature(consensusMessage)
				prePrepareMessage := &plum.XBFTRequest{
					Message:   consensusMessage,
					Signature: signature,
					Block:     p.D.CandidateBlocks[p.ID],
				}
				p.D.handleXBFT(prePrepareMessage)
				go SendAllExceptThisPeer(prePrepareMessage)
			}
		}
	}
}
