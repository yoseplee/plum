package peer

import (
	"log"
	"testing"
)

func TestPeer_selectionValue(t *testing.T) {
	for i := 0; i < 10; i++ {
		log.Println(GetInstance().selectionValue([]byte{}, GetInstance().ID))
	}
}

//func TestPeer_Selection(t *testing.T) {
//	log.Println("selection result:", GetInstance().Selection())
//}

func TestPeer_expectedCommitteeSize(t *testing.T) {
	want := 5.2
	if got := p.expectedCommitteeSize(); got != want {
		t.Errorf("invalid calculation of expected committee size. got: %v, want: %v", got, want)
	}
}

func TestPeer_minimumCommitteeSize(t *testing.T) {
	want := 4
	if got := p.minimumCommitteeSize(); got != want {
		t.Errorf("invalid calculation of the minimum size of committee. got: %v, want: %v", got, want)
	}
}

func TestCalcFaultyNodeSize(t *testing.T) {
	var want float64

	want = 1.0
	if got := calcFaultyNodeSize(4); got != want {
		t.Errorf("invalid calculation of faulty node size. got: %v, want: %v", got, want)
	}

	want = 1.3333333333333333
	if got := calcFaultyNodeSize(5); got != want {
		t.Errorf("invalid calculation of faulty node size. got: %v, want: %v", got, want)
	}

	want = 1.6666666666666667
	if got := calcFaultyNodeSize(6); got != want {
		t.Errorf("invalid calculation of faulty node size. got: %v, want: %v", got, want)
	}

	want = 2
	if got := calcFaultyNodeSize(7); got != want {
		t.Errorf("invalid calculation of faulty node size. got: %v, want: %v", got, want)
	}
}
