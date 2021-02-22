package peer

import (
	"github.com/yoseplee/plum/core/plum"
	"log"
	"sync"
	"time"
)

const (
	TimeoutPBFTNewRound    = 2 * time.Second
	TimeoutPBFTPrePrepare  = 2 * time.Second
	TimeoutPBFTPrepare     = 2 * time.Second
	TimeoutPBFTCommit      = 2 * time.Second
	TimeoutXBFTSelection   = 7 * time.Second
	TimeoutXBFTRoundChange = 10 * time.Second
	TimeoutXBFTPrePrepare  = 10 * time.Second
	TimeoutXBFTPrepare     = 10 * time.Second
	TimeoutXBFTCommit      = 10 * time.Second
)

var (
	kOnce  sync.Once
	keeper *Keeper
)

type Keeper struct {
	SetRound uint64
	SetPhase int32
	Timer    *time.Timer
	StopSig  chan bool
}

func GetKeeperInstance() *Keeper {
	kOnce.Do(func() {
		keeper = &Keeper{
			SetRound: 0,
			SetPhase: -1,
			Timer:    time.NewTimer(time.Microsecond),
			StopSig:  make(chan bool),
		}
		<-keeper.Timer.C //to prevent unexpected timeout handling in the run method()
		go keeper.run()
	})
	return keeper
}

func (k *Keeper) Set(setRound uint64, ph interface{}) {
	k.SetRound = setRound
	switch setPhase := ph.(type) {
	case plum.PBFTPhase:
		k.SetPhase = int32(setPhase)
		k.setPBFTTimer(setPhase)
	case plum.XBFTPhase:
		k.SetPhase = int32(setPhase)
		k.setXBFTTimer(setPhase)
	}
}

func (k *Keeper) Reset() {
	if !k.Timer.Stop() {
		return
	}
	k.setDefault()
}

func (k *Keeper) setDefault() {
	k.SetRound = 0
	k.SetPhase = -1
}

func (k *Keeper) run() {
	p := GetInstance()
	switch p.D.ConsensusType {
	case "PBFT":
	PEXIT:
		for {
			select {
			case <-k.Timer.C:
				//validation check
				//if round has been proceeded already, it doesn't have to handle new round
				if k.SetRound < p.ConsensusRound || k.SetPhase < int32(p.PBFTPhase) {
					k.setDefault()
					break
				}
				p.D.triggerPBFTRoundChange()
				k.setDefault()
			case <-k.StopSig:
				break PEXIT
			}
		}
	case "XBFT":
	XEXIT:
		for {
			select {
			case <-k.Timer.C:
				//validation check
				//if round has been proceeded already, it doesn't have to handle new round
				if k.SetRound < p.ConsensusRound || k.SetPhase < int32(p.XBFTPhase) {
					k.setDefault()
					break
				}
				log.Println("triggered at round:", k.SetRound, "| on:", plum.XBFTPhase(k.SetPhase))
				log.Println("peer's round:", p.ConsensusRound, "| on:", p.XBFTPhase)
				p.D.triggerXBFTRoundChange()
				k.setDefault()
			case <-k.StopSig:
				break XEXIT
			}
		}
	default:
		log.Fatalf("invalid consensus type at keeper run()")
	}
}

func (k *Keeper) stop() {
	k.StopSig <- true
}

func (k *Keeper) setPBFTTimer(phase plum.PBFTPhase) {
	switch phase {
	case plum.PBFTPhase_PBFTNewRound:
		k.Timer.Reset(TimeoutPBFTNewRound)
	case plum.PBFTPhase_PBFTPrePrepare:
		k.Timer.Reset(TimeoutPBFTPrePrepare)
	case plum.PBFTPhase_PBFTPrepare:
		k.Timer.Reset(TimeoutPBFTPrepare)
	case plum.PBFTPhase_PBFTCommit:
		k.Timer.Reset(TimeoutPBFTCommit)
	default:
		log.Println("failed to set time because of the invalid phase")
	}
}

func (k *Keeper) setXBFTTimer(phase plum.XBFTPhase) {
	switch phase {
	case plum.XBFTPhase_XBFTSelect:
		k.Timer.Reset(TimeoutXBFTSelection)
	case plum.XBFTPhase_XBFTRoundChange:
		k.Timer.Reset(TimeoutXBFTRoundChange)
	case plum.XBFTPhase_XBFTPrePrepare:
		k.Timer.Reset(TimeoutXBFTPrePrepare)
	case plum.XBFTPhase_XBFTPrepare:
		k.Timer.Reset(TimeoutXBFTPrepare)
	case plum.XBFTPhase_XBFTCommit:
		k.Timer.Reset(TimeoutXBFTCommit)
	default:
		log.Println("failed to set time because of the invalid phase")
	}
}
