package peer

import (
	"github.com/yoseplee/plum/core/plum"
	"log"
	"sort"
)

const (
	repIncreaseUnit = 0.01
	repDecreaseUnit = 0.09
)

//initReputation() loads reputation of peers. This can be changed to retrieve reputation record from data sources
func (p *peer) initReputation() {
	for _, a := range p.AddressBook {
		p.ReputationBook[a.PeerId] = 1.0
	}
}

//RepWeight() calculates node's portion of reputation
func (p peer) RepRatio(peerID uint32) float64 {
	repSum := p.RepSum()
	repHave := p.ReputationBook[peerID]
	return repHave / repSum
}

func (p peer) RepSum() float64 {
	var sum float64
	for _, v := range p.ReputationBook {
		sum += v
	}
	return sum
}

func (p *peer) RepMedian() float64 {
	var reps []float64
	for _, v := range p.ReputationBook {
		reps = append(reps, v)
	}
	sort.Float64s(reps)
	return reps[len(reps)/2]
}

func (p *peer) RepMedianRatio() float64 {
	return p.RepMedian() / p.RepSum()
}

//RepIncrease() linearly increases reputation of committee member
//when they successfully appended their candidate block
func (p *peer) RepIncrease(cms []*plum.CommitteeMembers) {
	p.mutex.Lock()
	for _, cm := range cms {
		p.ReputationBook[cm.PeerId] = repIncreaseUnit + p.ReputationBook[cm.PeerId]
	}
	p.mutex.Unlock()
}

func (p *peer) RepDecrease(cms []*plum.CommitteeMembers) {
	p.mutex.Lock()
	for _, cm := range cms {
		log.Printf("decrease reputation of %d: %v -> %v\n", cm.PeerId, p.ReputationBook[cm.PeerId], p.ReputationBook[cm.PeerId]*repDecreaseUnit)
		p.ReputationBook[cm.PeerId] = repDecreaseUnit * p.ReputationBook[cm.PeerId]
	}
	p.mutex.Unlock()
}
