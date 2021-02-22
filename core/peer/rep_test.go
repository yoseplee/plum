package peer

import (
	"github.com/yoseplee/plum/core/plum"
	"testing"
)

func TestPeer_initReputation(t *testing.T) {
	p := GetInstance()
	for k := range p.ReputationBook {
		p.ReputationBook[k] = 10.5
	}

	p.initReputation()
	want := 1.0
	for _, v := range p.ReputationBook {
		if v != want {
			t.Errorf("invalid init of reputation. want: %v, got: %v", want, v)
		}
	}
}

func TestPeer_RepRatio(t *testing.T) {
	p := GetInstance()
	//for equally distributed reputation
	want := 1.0 / float64(len(p.AddressBook))
	if got := p.RepRatio(p.ID); got != want {
		t.Errorf("invalid calculation of reputation weight. want: %v, got :%v", want, got)
	}
	for k := range p.ReputationBook {
		p.ReputationBook[k] = 10.5
	}
	if got := p.RepRatio(p.ID); got != want {
		t.Errorf("invalid calculation of reputation weight. want: %v, got :%v", want, got)
	}
}

func TestPeer_RepMedian(t *testing.T) {
	p := setDefaultReputation()
	want := 3.0
	if got := p.RepMedian(); got != want {
		t.Errorf("invalid calculation of replutation median. got: %v, want: %v", got, want)
	}
}

func TestPeer_RepMedianRatio(t *testing.T) {
	p := setDefaultReputation()
	want := 0.1111111111111111
	if got := p.RepMedianRatio(); got != want {
		t.Errorf("invalid calculation of reputation median ratio. got: %v, want: %v", got, want)
	}
}

func setDefaultReputation() *peer {
	p := GetInstance()
	p.ReputationBook[0] = 1.0
	p.ReputationBook[1] = 1.0
	p.ReputationBook[2] = 1.0
	p.ReputationBook[3] = 3.0
	p.ReputationBook[4] = 7.0
	p.ReputationBook[5] = 7.0
	p.ReputationBook[6] = 7.0
	return p
}

func TestPeer_RepIncrease(t *testing.T) {
	setDefaultReputation()
	cms := []*plum.CommitteeMembers{
		{PeerId: 0},
		{PeerId: 1},
		{PeerId: 2},
		{PeerId: 3},
	}

	before := make(map[uint32]float64)
	for k, _ := range cms {
		before[uint32(k)] = p.ReputationBook[uint32(k)]
	}

	p.RepIncrease(cms)

	for k, v := range before {
		before[k] = repIncreaseUnit + v
	}

	for k, v := range before {
		want := v
		if got := p.ReputationBook[k]; got != want {
			t.Errorf("invalid calculation of increasing reputation. got: %v, want: %v", got, want)
		}
	}
}

func TestPeer_RepDecrease(t *testing.T) {
	setDefaultReputation()
	cms := []*plum.CommitteeMembers{
		{PeerId: 0},
		{PeerId: 1},
		{PeerId: 2},
		{PeerId: 3},
	}

	before := make(map[uint32]float64)
	for k, _ := range cms {
		before[uint32(k)] = p.ReputationBook[uint32(k)]
	}

	p.RepDecrease(cms)

	for k, v := range before {
		before[k] = repDecreaseUnit * v
	}

	for k, v := range before {
		want := v
		if got := p.ReputationBook[k]; got != want {
			t.Errorf("invalid calculation of increasing reputation. got: %v, want: %v", got, want)
		}
	}
}
