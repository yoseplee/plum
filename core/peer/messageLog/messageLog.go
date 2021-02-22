package messageLog

import (
	"github.com/yoseplee/plum/core/plum"
	"log"
)

type LogManager interface {
	Store(interface{})
	Clear()
	Get(r uint64, ph interface{}) interface{}
}

type MessageLog struct {
	log []*plum.XBFTRequest
}

func (m *MessageLog) Store(i interface{}) {
	switch i.(type) {
	case *plum.XBFTRequest:
		m.log = append(m.log, i.(*plum.XBFTRequest))
	default:
		log.Panic("invalid type for MessageLog Store()")
	}
}

func (m *MessageLog) Clear() {
	m.log = nil
}

func (m *MessageLog) Get(r uint64, ph interface{}) interface{} {
	var prePrepares []*plum.XBFTRequest
	for i := len(m.log); i > 0; i-- {
		idx := i - 1
		l := m.log[idx]

		if l.GetMessage().GetRound() < r {
			break
		}

		if l.GetMessage().GetRound() == r && l.GetMessage().GetPhase() == ph.(plum.XBFTPhase) {
			prePrepares = append(prePrepares, l)
		}
	}
	return prePrepares
}
