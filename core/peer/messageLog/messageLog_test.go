package messageLog

import (
	"github.com/yoseplee/plum/core/plum"
	"testing"
)

func TestMessageLog_Store(t *testing.T) {
	var m LogManager
	m = &MessageLog{}
	for i := 0; i < 100; i++ {
		m.Store(&plum.XBFTRequest{
			Message: &plum.XBFTMessage{
				Phase: plum.XBFTPhase_XBFTSelect,
				Round: uint64(i),
			}},
		)
	}
}

func TestMessageLog_Get(t *testing.T) {
	var m LogManager
	m = &MessageLog{}
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(0),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(0),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(0),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(0),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(0),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(1),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(1),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(1),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrepare,
			Round: uint64(2),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(3),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(3),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTCommit,
			Round: uint64(4),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(5),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTPrePrepare,
			Round: uint64(6),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTSelect,
			Round: uint64(7),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTCommit,
			Round: uint64(8),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTCommit,
			Round: uint64(9),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTCommit,
			Round: uint64(9),
		}},
	)
	m.Store(&plum.XBFTRequest{
		Message: &plum.XBFTMessage{
			Phase: plum.XBFTPhase_XBFTCommit,
			Round: uint64(9),
		}},
	)

	var want int

	want = 5
	if got := len(m.Get(0, plum.XBFTPhase_XBFTSelect).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved selection messages. got: %v, want: %v", got, want)
	}

	want = 3
	if got := len(m.Get(1, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 2
	if got := len(m.Get(3, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 1
	if got := len(m.Get(6, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 0
	if got := len(m.Get(9, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 0
	if got := len(m.Get(8, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 0
	if got := len(m.Get(2, plum.XBFTPhase_XBFTPrePrepare).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved prepare messages. got: %v, want: %v", got, want)
	}

	want = 3
	if got := len(m.Get(9, plum.XBFTPhase_XBFTCommit).([]*plum.XBFTRequest)); got != want {
		t.Errorf("invalid number of retrieved commit messages. got: %v, want: %v", got, want)
	}
}
