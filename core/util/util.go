package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/plum"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

func GetExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network")
}

func MakeTarget(ipv4 string, port string) (string, error) {
	if len(port) == 0 {
		return "", fmt.Errorf("failed to make full target - port is missing")
	}

	if port[0] == byte(':') {
		port = strings.Trim(port, ":")
	}

	return fmt.Sprintf("%s:%s", ipv4, port), nil
}

func MakeString(in interface{}) string {
	var s string
	switch r := in.(type) {
	case *plum.PBFTRequest:
		s += fmt.Sprintf("[MSG: PBFTRequest] peer(%d), round(%d), phase(%s)", r.Message.GetPeerId(), r.Message.GetRound(), r.Message.GetPhase().String())
	case *plum.XBFTRequest:
		s += fmt.Sprintf("[MSG: XBFTRequest] peer(%d), round(%d), phase(%s), their_primary(%d)", r.Message.GetPeerId(), r.Message.GetRound(), r.Message.GetPhase().String(), r.GetMessage().GetPrimaryId())
	case *plum.PBFTResponse:
		switch rf := r.GetResult().(type) {
		case *plum.PBFTResponse_Msg:
			s += fmt.Sprintf("[MSG: PBFTResponse] status(%s), result(%s)", r.GetStatus(), rf.Msg)
		case *plum.PBFTResponse_Error:
			s += fmt.Sprintf("[MSG: PBFTResponse] status(%s), result(%s)", r.GetStatus(), rf.Error)
		}
	case *plum.Block:
		s += fmt.Sprintf("%s\n", "=== === === === === === ===")
		s += fmt.Sprintf("[ %d th ] %s\n", r.Header.Id, "Block")
		s += fmt.Sprintf("%s\n", "=== === === HEADER === === ===")
		s += fmt.Sprintf("%-15s | %d\n", "ID: ", r.Header.Id)
		s += fmt.Sprintf("%-15s | %s\n", "MerkleRoot: ", hex.EncodeToString(r.Header.MerkleRoot))
		s += fmt.Sprintf("%-15s | %s\n", "PrevBlockHash: ", hex.EncodeToString(r.Header.PrevBlockHash))
		s += fmt.Sprintf("%-15s | %s\n", "Time: ", r.Header.Time)
		s += fmt.Sprintf("%s\n", "=== === === BODY === === ===")
		s += fmt.Sprintf("%s\n", r.Body.MerkleTree.Root.String())
		//s += fmt.Sprintf("%s\n", "=== === === TXS === === ===")
		//for i, tx := range r.Body.Txs {
		//	s += fmt.Sprintf("%d | %v\n", i, tx)
		//}
	case *plum.MerkleNode:
		if r.L == nil && r.R == nil {
			return fmt.Sprintf("TERMINAL NODE (left = %v, right = %v, data = %s", nil, nil, hex.EncodeToString(r.D))
		} else {
			return fmt.Sprintf("NODE (left = %v,\nright = %v,\ndata = %s\n", r.L, r.R, hex.EncodeToString(r.D))
		}
	default:
		s = fmt.Sprintf("%s", in)
	}
	return s
}

func StoreNewGenesisBlock() {
	g := block.NewGenesisBlock()
	//marshal
	mg, err := proto.Marshal(g)
	if err != nil {
		log.Fatalln("could not marshal message properly")
	}

	//file store
	fErr := ioutil.WriteFile("/Users/yosep/Documents/github/plum/core/genesis.block", mg, os.FileMode(777))
	if fErr != nil {
		log.Fatalf("could not save genesis block properly: %v", fErr)
	}
}

func DebugMsg(msg ...string) {
	log.Println(fmt.Sprintf("[DEBUG] %s", msg))
}

func PrintCommitteeMembers(cms []*plum.CommitteeMembers) {
	var s string
	var r uint64
	if len(cms) == 0 {
		log.Println("EMPTY COMMITTEE MEMBERS")
		return
	}
	r = cms[0].GetRound()
	s += fmt.Sprintf("PEER ID | SELECTION VALUE\n")
	for _, cm := range cms {
		s += fmt.Sprintf("%7d | %v\n", cm.PeerId, cm.SelectionValue)
	}
	log.Println(fmt.Sprintf("\n=== COMMITTEE MEMBERS AT ROUND %d ===\n%s", r, s))
}
