package heap

import (
	"errors"
	"github.com/yoseplee/plum/core/plum"
	"sync"
)

type MinPBFTHeap struct {
	heap
}

func NewMinPBFTHeap() *MinPBFTHeap {
	mh := &MinPBFTHeap{}
	mh.Last = -1
	mh.mutex = &sync.Mutex{}
	return mh
}

func (mh *MinPBFTHeap) Push(m *plum.PBFTRequest) {
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

func (mh *MinPBFTHeap) Pop() (*plum.PBFTRequest, error) {

	mh.mutex.Lock()
	//when it is empty
	if mh.Empty() {
		mh.mutex.Unlock()
		return nil, errors.New("empty heap")
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
