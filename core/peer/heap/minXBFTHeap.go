package heap

import (
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"sync"
)

type MinXBFTHeap struct {
	Q     []*plum.XBFTRequest
	Last  int
	mutex *sync.Mutex
}

func NewMinXBFTHeap() *MinXBFTHeap {
	mh := &MinXBFTHeap{}
	mh.Last = -1
	mh.mutex = &sync.Mutex{}
	return mh
}

func (h MinXBFTHeap) String() string {
	var s string
	h.mutex.Lock()
	s += fmt.Sprintf("%s\n", "printing MinXBFTHeap")
	s += fmt.Sprintf("%s:%d | len(%d), cap(%d)\n", "# of nodes", h.GetLast(), len(h.Q), cap(h.Q))
	s += fmt.Sprintf("===== %s =====\n", "data")
	if h.Q == nil {
		s += fmt.Sprintf("%s\n", "empty MinXBFTHeap")
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

func (h MinXBFTHeap) GetLast() int {
	return h.Last
}

func (h MinXBFTHeap) Peek() (*plum.XBFTRequest, error) {
	if h.Empty() {
		return nil, errors.New("empty MinXBFTHeap")
	}
	return h.Q[0], nil
}

func (h MinXBFTHeap) Empty() bool {
	if h.Last == -1 || len(h.Q) == 0 {
		return true
	}
	return false
}

func (h MinXBFTHeap) ParentIdx(i int) int {
	return (i - 1) / 2
}

func (h MinXBFTHeap) LeftChildIdx(i int) int {
	return (i * 2) + 1
}

func (h MinXBFTHeap) RightChildIdx(i int) int {
	return (i * 2) + 2
}

func (h MinXBFTHeap) HasLeftChild(i int) bool {
	lci := h.LeftChildIdx(i)
	if lci > h.Last {
		return false
	}
	return true
}

func (h MinXBFTHeap) HasRightChild(i int) bool {
	rci := h.RightChildIdx(i)
	if rci > h.Last {
		return false
	}
	return true
}

func (mh *MinXBFTHeap) Push(m *plum.XBFTRequest) {

	//insert to the last of the tree
	mh.mutex.Lock()
	mh.Q = append(mh.Q, m)
	mh.Last++

	i := mh.Last

	for {
		n := mh.Q[i]
		if i == 0 {
			//reach to the top
			break
		}

		parentIdx := mh.ParentIdx(i)
		parent := mh.Q[parentIdx]
		if n.Message.GetRound() < parent.Message.GetRound() {
			//swap
			mh.Q[parentIdx] = n
			mh.Q[i] = parent
			i = parentIdx
		} else if n.Message.GetRound() == parent.Message.GetRound() && n.Message.GetPhase() < parent.Message.GetPhase() {
			//swap
			mh.Q[parentIdx] = n
			mh.Q[i] = parent
			i = parentIdx
		} else {
			break
		}
	}
	mh.mutex.Unlock()
}

func (mh *MinXBFTHeap) Pop() (*plum.XBFTRequest, error) {

	mh.mutex.Lock()
	//when it is empty
	if mh.Empty() {
		mh.mutex.Unlock()
		return nil, errors.New("empty MinXBFTHeap")
	}

	//pop root node
	m := mh.Q[0]

	//select the last node then set this as top node
	targetIdx := 0
	mh.Q[targetIdx] = mh.Q[mh.Last]
	mh.Q = mh.Q[:mh.Last] //throw away remained one
	mh.Last--

	//reform
	for {
		lchi := mh.LeftChildIdx(targetIdx)
		rchi := mh.RightChildIdx(targetIdx)

		//this is a terminal node
		if lchi > mh.Last {
			break
		}

		//when it has only left child
		if lchi == mh.Last {
			//compare
			if mh.Q[lchi].Message.GetRound() < mh.Q[targetIdx].Message.GetRound() {
				//swap with left
				tmp := mh.Q[lchi]
				mh.Q[lchi] = mh.Q[targetIdx]
				mh.Q[targetIdx] = tmp
				targetIdx = lchi
			} else if mh.Q[lchi].Message.GetRound() == mh.Q[targetIdx].Message.GetRound() && mh.Q[lchi].Message.GetPhase() < mh.Q[targetIdx].Message.GetPhase() {
				//tie break rule
				//swap with left
				tmp := mh.Q[lchi]
				mh.Q[lchi] = mh.Q[targetIdx]
				mh.Q[targetIdx] = tmp
				targetIdx = lchi
			}
			break
		} else {
			//when it has both children
			if mh.Q[lchi].Message.GetRound() < mh.Q[rchi].Message.GetRound() {
				if mh.Q[lchi].Message.GetRound() < mh.Q[targetIdx].Message.GetRound() {
					//swap with left
					tmp := mh.Q[lchi]
					mh.Q[lchi] = mh.Q[targetIdx]
					mh.Q[targetIdx] = tmp
					targetIdx = lchi
				} else {
					break
				}
			} else if mh.Q[lchi].Message.GetRound() > mh.Q[rchi].Message.GetRound() {
				if mh.Q[rchi].Message.GetRound() < mh.Q[targetIdx].Message.GetRound() {
					//swap with right
					tmp := mh.Q[rchi]
					mh.Q[rchi] = mh.Q[targetIdx]
					mh.Q[targetIdx] = tmp
					targetIdx = rchi
				} else {
					break
				}
			} else if mh.Q[lchi].Message.GetRound() == mh.Q[rchi].Message.GetRound() {
				//select left or right by phase
				if mh.Q[lchi].Message.GetPhase() <= mh.Q[rchi].Message.GetPhase() {
					//left child
					//round check
					if mh.Q[lchi].Message.GetRound() < mh.Q[targetIdx].Message.GetRound() {
						//swap with left
						tmp := mh.Q[lchi]
						mh.Q[lchi] = mh.Q[targetIdx]
						mh.Q[targetIdx] = tmp
						targetIdx = lchi
					} else if mh.Q[lchi].Message.GetRound() == mh.Q[targetIdx].Message.GetRound() && mh.Q[lchi].Message.GetPhase() < mh.Q[targetIdx].Message.GetPhase() {
						//swap with left
						tmp := mh.Q[lchi]
						mh.Q[lchi] = mh.Q[targetIdx]
						mh.Q[targetIdx] = tmp
						targetIdx = lchi
					} else {
						break
					}
				} else if mh.Q[lchi].Message.GetPhase() > mh.Q[rchi].Message.GetPhase() {
					//right child
					//round check
					if mh.Q[rchi].Message.GetRound() < mh.Q[targetIdx].Message.GetRound() {
						//swap with left
						tmp := mh.Q[rchi]
						mh.Q[rchi] = mh.Q[targetIdx]
						mh.Q[targetIdx] = tmp
						targetIdx = rchi
					} else if mh.Q[rchi].Message.GetRound() == mh.Q[targetIdx].Message.GetRound() && mh.Q[rchi].Message.GetPhase() < mh.Q[targetIdx].Message.GetPhase() {
						//swap with left
						tmp := mh.Q[rchi]
						mh.Q[rchi] = mh.Q[targetIdx]
						mh.Q[targetIdx] = tmp
						targetIdx = rchi
					} else {
						break
					}
				}
			}
		}
	}
	mh.mutex.Unlock()
	return m, nil
}
