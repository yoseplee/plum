package mq

import (
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"log"
	"math/rand"
	"testing"
)

func TestQueue_Pop(t *testing.T) {
	q := NewPBFTQueue()
	for i := 0; i < 100; i++ {
		m := &plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: uint64(i),
			},
		}
		q.Push(m)
	}

	for i := 0; i < 50; i++ {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}

	want := uint64(50)
	if got := q.GetN(); got != want {
		t.Errorf("queue pop failed to be empty: got: %v, want: %v", got, want)
	}

	for !q.Empty() {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}

	want = uint64(0)
	if got := q.GetN(); got != want {
		t.Errorf("queue pop failed to be empty: got: %v, want: %v", got, want)
	}

	for i := 0; i < 100; i++ {
		m := &plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: uint64(i),
			},
		}
		q.Push(m)
	}

	for !q.Empty() {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}

	want = uint64(0)
	if got := q.GetN(); got != want {
		t.Errorf("queue pop failed to be empty: got: %v, want: %v", got, want)
	}
}

func TestQueue_Peek(t *testing.T) {
	q := NewPBFTQueue()
	for i := 0; i < 100; i++ {
		m := &plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(i),
				Round: uint64(i),
			},
		}
		q.Push(m)
	}
	want := &node{D: &plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Round: 0,
			Phase: 0,
		},
	}}
	if got := q.Peek(); got.D.GetMessage().GetRound() != want.D.GetMessage().GetRound() && q.GetN() == uint64(100) {
		t.Errorf("invalid peek method")
	}
}

func TestQueue_Empty(t *testing.T) {
	q := NewPBFTQueue()
	var want bool
	want = true
	if got := q.Empty(); got != want {
		t.Errorf("queue is not empty")
	}

	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Round: 0,
			Phase: 0,
		},
	})
	want = false
	if got := q.Empty(); got != want {
		t.Errorf("queue is empty")
	}

	for i := 0; i < 100; i++ {
		m := &plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: uint64(i),
			},
		}
		q.Push(m)
	}

	for !q.Empty() {
		_, err := q.Pop()
		if err != nil {
			continue
		}
	}

	want = true
	if got := q.Empty(); got != want {
		t.Errorf("queue is not empty")
	}
}

func TestQueue_GetN(t *testing.T) {
	q := NewPBFTQueue()
	for i := 0; i < 100; i++ {
		m := &plum.PBFTRequest{
			Message: &plum.PBFTMessage{
				Phase: plum.PBFTPhase(rand.Intn(4)),
				Round: uint64(i),
			},
		}
		q.Push(m)
	}
	want := uint64(100)
	if got := q.GetN(); want != got {
		t.Errorf("invalid queue counting")
	}
}

func TestQueue_PushPop(t *testing.T) {
	q := NewPBFTQueue()
	maxPushIter := 1000

	pushDone := make(chan struct{})

	go func() {
		for i := 0; i < maxPushIter; i++ {
			m := &plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i),
					Round: uint64(i),
				},
			}
			q.Push(m)
		}
		close(pushDone)
	}()

	go func() {
		for {
			_, err := q.Pop()
			if err != nil {
				continue
			}
		}
	}()

	<-pushDone

	want := uint64(0)
	if got := q.GetN(); got > want+50 {
		t.Errorf("failed to pop goroutine, want: %v, got: %v", want, got)
	}
}

func TestQueue_PushPop_Goroutine_Safe(t *testing.T) {
	q := NewPBFTQueue()

	//insert for 1000 times
	done1 := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			q.Push(&plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i),
					Round: uint64(i),
				},
			})
		}
		close(done1)
	}()

	//insert for 1000 times
	done2 := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			q.Push(&plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i),
					Round: uint64(i),
				},
			})
		}
		close(done2)
	}()

	//delete for 1000 times
	deleteDone := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			_, err := q.Pop()
			if err != nil {
				continue
			}
		}
		close(deleteDone)
	}()

	//insert for 1000 times
	done3 := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			q.Push(&plum.PBFTRequest{
				Message: &plum.PBFTMessage{
					Phase: plum.PBFTPhase(i),
					Round: uint64(i),
				},
			})
		}
		close(done3)
	}()

	<-done1
	<-done2
	<-done3
	<-deleteDone

	log.Println(q.GetN())
}

func TestQueue_String(t *testing.T) {
	q := NewPBFTQueue()

	log.Println(q.GetN(), q.String())
}

func TestQueue_Push_Goroutines_Concurrency(t *testing.T) {
	q := NewPBFTQueue()

	done := make(chan bool, 5)
	for c := 0; c < 5; c++ {
		c := c
		go func() {
			iter := 1000
			base := c * iter
			for i := base; i < base+iter; i++ {
				q.Push(&plum.PBFTRequest{
					Message: &plum.PBFTMessage{
						Phase: plum.PBFTPhase(i),
						Round: uint64(i),
					},
				})
			}
			done <- true
		}()
	}

	var counter int
	for d := range done {
		if d == true {
			counter++
		}
		if counter == 5 {
			break
		}
	}
	log.Println(q.n)
}

func TestQueue_Push_Pop_Alternatively(t *testing.T) {
	q := NewPBFTQueue()

	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase(2),
			Round: uint64(2),
		},
	})
	r, err := q.Pop()
	if err != nil {
		log.Printf("%v", err)
	}
	q.Push(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase: plum.PBFTPhase(1),
			Round: uint64(1),
		},
	})
	r, err = q.Pop()
	if err != nil {
		log.Printf("%v", err)
	}
	log.Println(util.MakeString(r.D))
}
