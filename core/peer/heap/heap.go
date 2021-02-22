package heap

import (
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"sync"
)

type heap struct {
	Q     []*plum.PBFTRequest
	Last  int
	mutex *sync.Mutex
}

func NewHeap() *heap {
	return &heap{
		Q:     make([]*plum.PBFTRequest, 10, 1000),
		Last:  -1,
		mutex: &sync.Mutex{},
	}
}
func (h heap) String() string {
	var s string
	h.mutex.Lock()
	s += fmt.Sprintf("%s\n", "printing heap")
	s += fmt.Sprintf("%s:%d | len(%d), cap(%d)\n", "# of nodes", h.GetLast(), len(h.Q), cap(h.Q))
	s += fmt.Sprintf("===== %s =====\n", "data")
	if h.Q == nil {
		s += fmt.Sprintf("%s\n", "empty heap")
	} else {
		for i, d := range h.Q {
			s += fmt.Sprintf("idx(%d) | %s\n", i, util.MakeString(d))
			if i > 10 {
				s += fmt.Sprintf("%s(%d) %s\n", "...", h.GetLast()-9, "more")
				break
			}
		}
	}
	h.mutex.Unlock()
	s += fmt.Sprintf("====== == ======\n")
	return s
}

func (h heap) GetLast() int {
	return h.Last
}

func (h heap) Peek() (*plum.PBFTRequest, error) {
	if h.Empty() {
		return nil, errors.New("empty heap")
	}
	return h.Q[0], nil
}

func (h heap) Empty() bool {
	if h.Last == -1 || len(h.Q) == 0 {
		return true
	}
	return false
}

func (h heap) ParentIdx(i int) int {
	return (i - 1) / 2
}

func (h heap) LeftChildIdx(i int) int {
	return (i * 2) + 1
}

func (h heap) RightChildIdx(i int) int {
	return (i * 2) + 2
}

func (h heap) HasLeftChild(i int) bool {
	lci := h.LeftChildIdx(i)
	if lci > h.Last {
		return false
	}
	return true
}

func (h heap) HasRightChild(i int) bool {
	rci := h.RightChildIdx(i)
	if rci > h.Last {
		return false
	}
	return true
}
