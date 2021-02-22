package peer

import (
	"github.com/yoseplee/plum/core/plum"
	"testing"
	"time"
)

var k *Keeper

func TestGetKeeperInstance(t *testing.T) {
	i := GetKeeperInstance()
	if i.StopSig == nil {
		t.Errorf("invalid instance")
	}
}

func TestKeeper_Set(t *testing.T) {
	var want uint64
	p := GetInstance()

	//case1: timer expires
	k.Set(0, plum.PBFTPhase_PBFTNewRound)
	<-time.After(TimeoutPBFTNewRound + (time.Millisecond * 300)) //expect that it expires
	//i.e. round should be increased
	want = 1
	if got := p.ConsensusRound; got != want {
		t.Errorf("invalid calculation of round when the timer expired: got :%v, want: %v", got, want)
	}

	//case2: when the timer stopped
	k.Set(1, plum.PBFTPhase_PBFTPrePrepare)
	k.Reset()
	want = 1
	if got := p.ConsensusRound; got != want {
		t.Errorf("invalid calculation of round when the timer stopped")
	}

	//case3: timer expires but only phase proceeded
	k.Set(1, plum.PBFTPhase_PBFTPrePrepare)
	<-time.After(time.Millisecond * 50)
	p.PBFTPhase = plum.PBFTPhase_PBFTCommit

	//expires after, expect that it discards
	<-time.After(TimeoutPBFTPrePrepare + (time.Millisecond * 300))
	want = 1
	if got := p.ConsensusRound; got != want {
		t.Errorf("invalid calculation of round when the timer stopped")
	}
}

func TestKeeper_Set2(t *testing.T) {
	k.setPBFTTimer(-1)
}

func TestKeeper_Clear(t *testing.T) {
	k.Set(0, plum.PBFTPhase_PBFTCommit)
	k.Reset()
	if k.SetPhase != -1 {
		t.Errorf("the keeper didn't clear timer variables properly")
	}

	//try clear after it expired
	k.Set(1, plum.PBFTPhase_PBFTCommit)
	<-time.After(time.Millisecond * 1100)
	k.Reset()
	if k.SetPhase != -1 {
		t.Errorf("the keeper didn't clear timer variables properly")
	}
}
