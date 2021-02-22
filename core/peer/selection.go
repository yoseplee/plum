package peer

import (
	"log"
	"math/big"
)

const selectionThreshold float64 = 0.2
const F float64 = 10.0

//Selection() decides that this node is selected or not based on selection value
//if the selection value is larger than selection threshold which denoted as selectionThreshold
//then it knows that it is selected
func Selection(selectionValue float64) bool {
	if selectionValue > selectionThreshold {
		return true
	}
	return false
}

func (p *peer) selectionValue(vrfHash []byte, peerID uint32) float64 {
	gamma := p.getHashRatio(vrfHash)
	rho := p.RepRatio(peerID)
	return gamma + F*rho
}

func (p *peer) getHashRatio(verifiableHash []byte) float64 {
	t := &big.Int{}
	t.SetBytes(verifiableHash)

	precision := uint(8 * (len(verifiableHash) + 1))
	max, b, err := big.ParseFloat("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Fatal("failed to parse big float constant for selection")
	}

	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	r := big.Float{}
	ratio, _ := r.Quo(&h, max).Float64()

	return ratio
}

func (p *peer) expectedCommitteeSize() float64 {
	n := len(p.AddressBook)
	return float64(n) * (1 - (selectionThreshold - F*p.RepMedianRatio()))
}

func (p *peer) minimumCommitteeSize() int {
	//var minimumCommitteeSize int
	//minimumCommitteeSize = 3*int(math.Floor(calcFaultyNodeSize(p.expectedCommitteeSize()))) + 1
	//if minimumCommitteeSize < 4 {
	//	minimumCommitteeSize = 4
	//}
	//return minimumCommitteeSize
	return 4
}

func calcFaultyNodeSize(n interface{}) float64 {
	var r float64
	switch k := n.(type) {
	case int:
		r = float64(k-1) / 3.0
	case int8:
		r = float64(k-1) / 3.0
	case int32:
		r = float64(k-1) / 3.0
	case int64:
		r = float64(k-1) / 3.0
	case float32:
		r = float64(k-1) / 3.0
	case float64:
		r = (k - 1) / 3.0
	default:
		log.Panic("invalid type for calculating faulty node size")
	}
	return r
}
