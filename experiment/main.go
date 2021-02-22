package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

var (
	goal = flag.Int("goal", 100, "set the round goal")
	qty  = flag.Int("qty", 10, "set the number of nodes of the experiment")
	tick int
)

func main() {
	flag.Parse()
	log.Printf("Starting Experiment with %d nodes up to %d rounds\n", *qty, *goal)

	var clientConns []*grpc.ClientConn
	var farmerClients []plum.FarmerClient
	var peerStates map[uint32][]*plum.PeerState
	var rwMutex *sync.RWMutex

	peerStates = make(map[uint32][]*plum.PeerState)
	rwMutex = new(sync.RWMutex)

	// dial connection
	profile := loadProfile()

	clientConns = makeConnectionAll(profile)
	defer func() {
		for _, c := range clientConns {
			c.Close()
		}
	}()

	// make connection as a peer client
	farmerClients = makeFarmerClients(clientConns)

	// get all the state for each 1 sec, appending to the peerStates
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		currentRound := 0
		prevRound := 0
		for {
			for _, fc := range farmerClients {
				ctx, _ := context.WithTimeout(context.Background(), time.Second)

				r, err := fc.GetPeerState(ctx, &plum.Empty{})
				if err != nil {
					log.Printf("could not get the state of peer: %v", err)
					continue
				}
				currentRound = int(r.ConsensusRound)

				// start from round 1 or greater
				if currentRound == prevRound {
					break
				}
				rwMutex.Lock()
				peerStates[r.GetId()] = append(peerStates[r.GetId()], r)
				rwMutex.Unlock()

			}
			// start from round 1 or greater
			if currentRound == prevRound {
				continue
			}

			log.Printf("tick %d | round %d\n", tick, currentRound)
			prevRound = currentRound
			tick++
			<-time.After(time.Millisecond * 100)
			if currentRound >= *goal {
				cancel()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("experiment done...")
				return
			default:
				log.Println("running experiment...")
				<-time.After(time.Millisecond * 500)
			}
		}
	}()

	<-ctx.Done()

	//analyze
	blockHeightByTick := make(map[uint32][]uint64)
	roundByTick := make(map[uint32][]uint64)
	selectedCountByTick := make(map[uint32][]uint64)
	selectedSumByForAll := make(map[uint32]uint64)
	selectedCountAverageForAll := make(map[uint32]float64)
	tentativeSelectedCountByTick := make(map[uint32][]uint64)
	reputationByTick := make(map[uint32][]float64)
	reputationAverageForAll := make(map[uint32]float64)
	reputationSumForAll := make(map[uint32]float64)

	for k, v := range peerStates {
		for _, record := range v {
			roundByTick[k] = append(roundByTick[k], record.ConsensusRound)
			blockHeightByTick[k] = append(blockHeightByTick[k], record.BlockHeight)
			selectedCountByTick[k] = append(selectedCountByTick[k], record.SelectedCount)
			tentativeSelectedCountByTick[k] = append(tentativeSelectedCountByTick[k], record.TentativeSelectedCount)
			reputationByTick[k] = append(reputationByTick[k], record.Reputation)
		}
	}

	//calc reputation sum
	for k, v := range reputationByTick {
		var sum float64
		for _, rep := range v {
			sum += rep
		}
		reputationSumForAll[k] = sum
	}

	//calc reputation average
	for k, v := range reputationSumForAll {
		reputationAverageForAll[k] = v / float64(tick)
	}

	//calc selected count sum
	for k, v := range selectedCountByTick {
		var sum uint64
		for _, sc := range v {
			sum += sc
		}
		selectedSumByForAll[k] = sum
	}

	//calc selected count average
	for k, v := range selectedSumByForAll {
		selectedCountAverageForAll[k] = float64(v) / float64(tick)
	}

	for k, v := range selectedCountByTick {
		var result string
		result += fmt.Sprintf("# id round tick selectedCount tentative_SC Reputation BlockHeight\n")
		for i, d := range v {
			result += fmt.Sprintf("%d %d %d %v %v %v %v\n", k, roundByTick[k][i], i, d, tentativeSelectedCountByTick[k][i], reputationByTick[k][i], blockHeightByTick[k][i])
		}
		//file1, _ := os.Create(fmt.Sprintf("result_%d.txt", k))
		//fmt.Fprint(file1, result)
		fileName := fmt.Sprintf("result_%d.txt", k)
		ioutil.WriteFile(fileName, []byte(result), 0644)
	}

	//calc tps
	var avgsRecord string
	avgsRecord += "#id tps avgReputation avgSelectedCount\n"
	for k, v := range blockHeightByTick {
		tpsBase := v[0]
		tpsEnd := v[len(v)-1]
		tps := (float64(tpsEnd-tpsBase) / float64(tick)) * 10
		avgsRecord += fmt.Sprintf("%d %v %v %v\n", k, tps, reputationAverageForAll[k], selectedCountAverageForAll[k])
	}

	fileName := "averages.txt"
	ioutil.WriteFile(fileName, []byte(avgsRecord), 0644)
}

func formatPeerState(p *plum.PeerState) string {
	var s string

	t, err := util.MakeTarget(p.GetIpv4(), p.GetPort())
	if err != nil {
		log.Println("could not make target, ", err)
	}

	s += fmt.Sprintf("\n| %s |\n", "================ PEER ================")
	s += fmt.Sprintf("|%-17s| %-20d |\n", "ID", p.GetId())
	s += fmt.Sprintf("|%-17s| %-20s |\n", "Address", t)
	s += fmt.Sprintf("|%-17s| %-20s |\n", "Role", p.Role.String())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Round", p.GetConsensusRound())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Current Primary", p.GetCurrentPrimary())
	s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus Phase", p.GetConsensusPhase().String())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[Prepare]", p.Vote[int32(plum.PBFTPhase_PBFTPrepare)])
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[Commit]", p.Vote[int32(plum.PBFTPhase_PBFTPrepare)])
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Vote[RoundChange]", p.Vote[int32(plum.PBFTPhase_PBFTRoundChange)])
	s += fmt.Sprintf("|%-17s| %-20s |\n", "Consensus State", p.GetConsensusState().String())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Block Height", p.GetBlockHeight())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Queue Length", p.GetQueueLength())
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Heap Length", p.GetHeapLength())
	s += fmt.Sprintf("|%-17s| %-20f |\n", "Reputation", p.Reputation)
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Selected Count", p.SelectedCount)
	s += fmt.Sprintf("|%-17s| %-20d |\n", "Tentative SC", p.TentativeSelectedCount)

	return s
}

func makeFarmerClients(clientConns []*grpc.ClientConn) []plum.FarmerClient {
	var peerClients []plum.FarmerClient
	for _, c := range clientConns {
		peerClients = append(peerClients, plum.NewFarmerClient(c))
	}
	return peerClients
}

func makeConnectionAll(profile []string) []*grpc.ClientConn {
	var conns []*grpc.ClientConn
	for _, pr := range profile {
		conn, err := grpc.Dial(pr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("failed to connect: %v", err)
		}
		conns = append(conns, conn)
	}
	return conns
}

func loadProfile() []string {
	var p []string
	p = append(p, "127.0.0.1:50051")
	p = append(p, "127.0.0.1:50061")
	p = append(p, "127.0.0.1:50071")
	p = append(p, "127.0.0.1:50081") //4
	p = append(p, "127.0.0.1:50091")
	p = append(p, "127.0.0.1:50101")
	p = append(p, "127.0.0.1:50111") //7
	p = append(p, "127.0.0.1:50121")
	p = append(p, "127.0.0.1:50131")
	p = append(p, "127.0.0.1:50141") //10
	p = append(p, "127.0.0.1:50151")
	p = append(p, "127.0.0.1:50161")
	p = append(p, "127.0.0.1:50171") //13
	p = append(p, "127.0.0.1:50181")
	p = append(p, "127.0.0.1:50191")
	p = append(p, "127.0.0.1:50201") //16
	p = append(p, "127.0.0.1:50211")
	p = append(p, "127.0.0.1:50221")
	p = append(p, "127.0.0.1:50231") //19
	p = append(p, "127.0.0.1:50241")
	p = append(p, "127.0.0.1:50251")
	p = append(p, "127.0.0.1:50261") //22
	return p[:*qty]
}
