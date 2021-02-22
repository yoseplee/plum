package heap

import (
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"math/rand"
	"testing"
)

func TestHeap_GetParent(t *testing.T) {
	h := NewHeap()
	var want int

	want = 0
	if got := h.ParentIdx(0); got != want {
		t.Errorf("failed to calculate index of parent")
	}

	want = 0
	if got := h.ParentIdx(1); got != want {
		t.Errorf("failed to calculate index of parent")
	}
	if got := h.ParentIdx(2); got != want {
		t.Errorf("failed to calculate index of parent")
	}

	want = 1
	if got := h.ParentIdx(3); got != want {
		t.Errorf("failed to calculate index of parent")
	}
	if got := h.ParentIdx(4); got != want {
		t.Errorf("failed to calculate index of parent")
	}

	want = 2
	if got := h.ParentIdx(5); got != want {
		t.Errorf("failed to calculate index of parent")
	}
	if got := h.ParentIdx(6); got != want {
		t.Errorf("failed to calculate index of parent")
	}
}

func TestHeap_LeftChildIdx(t *testing.T) {
	h := NewHeap()
	var want int
	want = 1
	if got := h.LeftChildIdx(0); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}

	want = 3
	if got := h.LeftChildIdx(1); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}

	want = 5
	if got := h.LeftChildIdx(2); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}
}

func TestHeap_RightChildIdx(t *testing.T) {
	h := NewHeap()
	var want int
	want = 2
	if got := h.RightChildIdx(0); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}

	want = 4
	if got := h.RightChildIdx(1); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}

	want = 6
	if got := h.RightChildIdx(2); got != want {
		t.Errorf("invalid left child. want: %v, got: %v", want, got)
	}
}

func TestMinHeap_Push(t *testing.T) {
	tSet := []uint64{13, 88, 112, 35, 17, 76}
	aSet := []uint64{13, 17, 76, 88, 35, 112}
	mh := NewMinPBFTHeap()

	for _, d := range tSet {
		mh.Push(&plum.PBFTRequest{Message: &plum.PBFTMessage{Round: d}})
	}
	//compare
	for i, d := range mh.Q {
		if got := d.Message.GetRound(); got != aSet[i] {
			t.Errorf("invalid tree adjustment after insertion. want: %v, got: %v", aSet[i], got)
		}
	}
}

func TestMinHeap_Delete(t *testing.T) {
	var want uint64
	var got *plum.PBFTRequest
	var err error

	tSet := []uint64{13, 88, 112, 35, 17, 76}
	aSet := []uint64{17, 35, 76, 88, 112}
	mh := NewMinPBFTHeap()

	for _, d := range tSet {
		mh.Push(&plum.PBFTRequest{Message: &plum.PBFTMessage{Round: d}})
	}

	want = 13
	got, err = mh.Pop()
	if err != nil {
		t.Errorf("could not delete under given condtition: %v", err)
	}
	if got.Message.GetRound() != want {
		t.Errorf("invalid deletion. want: %v, got: %v", want, got)
	}

	//compare
	for i, d := range mh.Q {
		if got := d.Message.GetRound(); got != aSet[i] {
			t.Errorf("invalid tree adjustment after deletion. want: %v, got: %v", aSet[i], got)
		}
	}
}

func TestMinHeap_Push_Delete(t *testing.T) {
	var want uint64
	var got *plum.PBFTRequest
	var err error

	tSet := []uint64{13, 88, 112, 35, 17, 76}
	aSet := []uint64{17, 35, 112, 88, 76}
	mh := NewMinPBFTHeap()

	for _, d := range tSet {
		mh.Push(&plum.PBFTRequest{Message: &plum.PBFTMessage{Round: d}})
	}

	want = 13
	got, err = mh.Pop()
	if err != nil {
		t.Errorf("could not delete under given condtition: %v", err)
	}
	if got.Message.GetRound() != want {
		t.Errorf("invalid deletion. want: %v, got: %v", want, got)
	}

	want = 17
	got, err = mh.Pop()
	if err != nil {
		t.Errorf("could not delete under given condtition: %v", err)
	}
	if got.Message.GetRound() != want {
		t.Errorf("invalid deletion. want: %v, got: %v", want, got)
	}

	want = 35
	got, err = mh.Pop()
	if err != nil {
		t.Errorf("could not delete under given condtition: %v", err)
	}
	if got.Message.GetRound() != want {
		t.Errorf("invalid deletion. want: %v, got: %v", want, got)
	}

	mh.Push(&plum.PBFTRequest{Message: &plum.PBFTMessage{Round: 17}})
	mh.Push(&plum.PBFTRequest{Message: &plum.PBFTMessage{Round: 35}})

	//compare
	for i, d := range mh.Q {
		if got := d.Message.GetRound(); got != aSet[i] {
			t.Errorf("invalid tree adjustment after deletion. want: %v, got: %v", aSet[i], got)
		}
	}
}

func TestMinHeap_Push_Delete_Massive(t *testing.T) {
	q := NewMinPBFTHeap()
	for i := 0; i < 10; i++ {
		q.Push(&plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(i),
				Round: uint64(i),
			},
		})
	}

	for i := 0; i < 10; i++ {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}

	for i := 0; i < 5000; i++ {
		if q == nil {
			break
		}
		q.Push(&plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(i),
				Round: uint64(i),
			},
		})
	}

	for i := 0; i < 5000; i++ {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}
}

func TestMinHeap_PushAndCount(t *testing.T) {
	targetCount := 0
	maxIter := 10000
	targetRound := 0

	mh := NewMinPBFTHeap()
	for i := 0; i < maxIter; i++ {
		round := uint64(rand.Intn(100))
		if round == uint64(targetRound) {
			targetCount++
		}
		mh.Push(&plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: round,
			},
		})
	}

	var want int
	want = (maxIter - 1) - targetCount

	if got := mh.Last - targetCount; got != want {
		t.Errorf("mismatch between heap last and counted last: got: %v, want: %v", got, want)
	}

	want = maxIter - targetCount
	if got := len(mh.Q) - targetCount; got != want {
		t.Errorf("mismatch between heap len and counted len: got: %v, want: %v", got, want)
	}
}

func TestMinHeap_Tie_Break_Rule_Case_1(t *testing.T) {
	q := NewMinPBFTHeap()
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 2,
			Round: 6,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 2,
			Round: 6,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 3,
			Round: 6,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 1,
			Round: 6,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 0,
			Round: 6,
		},
	})
	got, err := q.Pop()
	if err != nil {
		t.Errorf("could not pop: %v", err)
	}

	want := &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 0,
			Round: 6,
		},
	}
	if got.Message.GetRound() != want.Message.GetRound() && got.Message.GetPhase() != want.Message.GetPhase() {
		t.Errorf("invalid pop. want: %s, got: %s", util.MakeString(want), util.MakeString(got))
	}

	q1 := NewMinPBFTHeap()
	for i := 0; i < 100; i++ {
		q1.Push(&plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(3)),
				Round: uint64(rand.Intn(10)),
			},
		})
	}

	for i := 0; i < 5; i++ {
		r, err := q1.Pop()
		if err != nil {
			t.Errorf("could not pop: %v", err)
		}
		got = r
	}

	//statistically, almost all the case will be true
	want = &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 0,
			Round: 0,
		},
	}
	if got.Message.GetRound() != want.Message.GetRound() && got.Message.GetPhase() != want.Message.GetPhase() {
		t.Errorf("invalid pop. want: %s, got: %s", util.MakeString(want), util.MakeString(got))
	}
}

func TestMinHeap_Tie_Break_Rule_Case_2(t *testing.T) {
	q := NewMinPBFTHeap()
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrePrepare,
			Round: uint64(13),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTCommit,
			Round: uint64(13),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrepare,
			Round: uint64(13),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrePrepare,
			Round: uint64(14),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTCommit,
			Round: uint64(13),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTCommit,
			Round: uint64(13),
		},
	})
	q.Q = append(q.Q, &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrepare,
			Round: uint64(14),
		},
	})
	q.Last = 6

	_, err := q.Pop()
	if err != nil {
		t.Errorf("could not pop: %v", err)
	}

	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrePrepare,
			Round: uint64(13),
		},
	})

	got, err := q.Pop()
	if err != nil {
		t.Errorf("could not pop: %v", err)
	}

	want := &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase_PBFTPrePrepare,
			Round: uint64(13),
		},
	}

	if got.Message.GetPhase() != want.Message.GetPhase() || got.Message.GetRound() != want.Message.GetRound() {
		t.Errorf("invalid tree shape")
	}
}

func TestHeap_Peek(t *testing.T) {
	q := NewMinPBFTHeap()
	top, hErr := q.Peek()
	if hErr == nil {
		t.Errorf("invalid peek of the heap. the heap is empty but didn't emit error")
	}

	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 2,
			Round: 6,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 2,
			Round: 1,
		},
	})
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: 3,
			Round: 3,
		},
	})
	top, hErr = q.Peek()
	if hErr != nil {
		t.Errorf("could not peek: %v", hErr)
	}
	want := 1
	if got := top.GetMessage().GetRound(); got != uint64(want) {
		t.Errorf("invalid peek from the heap. want: %v got: %v", want, got)
	}
	_, err := q.Pop()
	if err != nil {
		t.Errorf("could not pop: %v", err)
	}

	top, hErr = q.Peek()
	if hErr != nil {
		t.Errorf("could not peek: %v", hErr)
	}
	want = 3
	if got := top.GetMessage().GetRound(); got != uint64(want) {
		t.Errorf("invalid peek from the heap. want: %v got: %v", want, got)
	}
}
