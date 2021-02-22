package mq

import (
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"sync"
)

type node struct {
	D    *plum.PBFTRequest
	next *node
}

type Queue struct {
	top   *node
	last  *node
	n     uint64
	mutex *sync.Mutex
}

func NewPBFTQueue() *Queue {
	return &Queue{
		mutex: &sync.Mutex{},
	}
}

func (q Queue) String() string {
	var s string
	s = ""
	q.mutex.Lock()
	if q.top == nil {
		s += fmt.Sprintf("%s", "empty queue")
		q.mutex.Unlock()
		return s
	} else {
		var i int
		for t := q.top; t != nil; t = t.next {
			s += fmt.Sprintf("%s, ", util.MakeString(t.D))
			if i > 10 {
				s += fmt.Sprintf("%s(%d) %s\n", "...", q.n-10, "more")
				break
			}
			i++
		}
	}
	q.mutex.Unlock()
	return s
}

func (q *Queue) Push(d *plum.PBFTRequest) {
	tn := &node{D: d}

	//when Queue is empty
	q.mutex.Lock()
	if q.Empty() {
		q.top = tn
		q.last = tn
		q.n++
		q.mutex.Unlock()
		return
	}

	//when Queue is not empty
	q.last.next = tn
	q.last = tn
	q.n++
	q.mutex.Unlock()
}

func (q *Queue) Pop() (*node, error) {

	q.mutex.Lock()
	//when Queue is empty
	if q.Empty() {
		q.mutex.Unlock()
		return nil, errors.New("empty queue")
	}

	//when Queue is not empty
	t := q.top
	q.top = q.top.next
	q.n--
	if q.n == 0 || q.top == nil {
		q.top = nil
		q.last = nil
	}
	q.mutex.Unlock()
	return t, nil
}

func (q *Queue) Peek() *node {
	//when Queue is empty
	if q.Empty() {
		return nil
	}
	q.mutex.Lock()

	q.mutex.Unlock()
	return q.top
}

func (q *Queue) Empty() bool {
	if q.n == 0 || q.top == nil {
		return true
	}
	return false
}

func (q *Queue) GetN() uint64 {
	return q.n
}
