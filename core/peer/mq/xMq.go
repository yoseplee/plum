package mq

import (
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"sync"
)

type xNode struct {
	D    *plum.XBFTRequest
	next *xNode
}

type XQueue struct {
	top   *xNode
	last  *xNode
	n     uint64
	mutex *sync.Mutex
}

func NewXBFTQueue() *XQueue {
	return &XQueue{
		mutex: &sync.Mutex{},
	}
}

func (q XQueue) String() string {
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

func (q *XQueue) Push(d *plum.XBFTRequest) {
	tn := &xNode{D: d}

	//when XQueue is empty
	q.mutex.Lock()
	if q.Empty() {
		q.top = tn
		q.last = tn
		q.n++
		q.mutex.Unlock()
		return
	}

	//when XQueue is not empty
	q.last.next = tn
	q.last = tn
	q.n++
	q.mutex.Unlock()
}

func (q *XQueue) Pop() (*xNode, error) {

	q.mutex.Lock()
	//when XQueue is empty
	if q.Empty() {
		q.mutex.Unlock()
		return nil, errors.New("empty queue")
	}

	//when XQueue is not empty
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

func (q *XQueue) Peek() *xNode {
	//when XQueue is empty
	if q.Empty() {
		return nil
	}
	q.mutex.Lock()

	q.mutex.Unlock()
	return q.top
}

func (q *XQueue) Empty() bool {
	if q.n == 0 || q.top == nil {
		return true
	}
	return false
}

func (q *XQueue) GetN() uint64 {
	return q.n
}
