package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/joho/godotenv"
	"github.com/yoseplee/plum/core/ledger"
	"github.com/yoseplee/plum/core/ledger/block"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"github.com/yoseplee/plum/core/util/path"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var localFlag = flag.Bool("local", false, "option for executing in local or distributed environment")
var targetFlag = flag.String("target", "", "target peer to send a message")
var iterFlag = flag.Int("iter", 1, "set how many times to iterate")
var roundFlag = flag.Uint64("round", 0, "set start(base) round on consensus")
var speedFlag = flag.Uint("speed", 0, "set speed to send new request to a peer, max: 3")
var operationFlag = flag.String("o", "", "operation to execute: triggerConsensus / getPeerState")

func main() {
	var conn *grpc.ClientConn
	var err error
	var privateKey ed25519.PrivateKey

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	flag.Parse()
	profile := loadProfile()
	privateKey = setPrivateKey(err, privateKey)
	latency := setLatency()

	conn, err = makeConnection(conn, err, profile)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	nextBlock, nextBlockDigest := setInitialBlockState()

	farmerClient := plum.NewFarmerClient(conn)
	peerClient := plum.NewPeerClient(conn)

	switch *operationFlag {
	case "triggerConsensus":
		handleTriggerConsensus(conn, latency, nextBlockDigest, privateKey, nextBlock)
	case "getPeerState":
		handleGetPeerState(farmerClient, latency)
	case "getPeerStateStream":
		handleGetPeerStateStream(farmerClient, latency)
	case "tps":
		calculateTps(farmerClient, latency)
	case "ping":
		handlePing(peerClient)
	default:
		log.Println("invalid operation")
	}
}

func handlePing(c plum.PeerClient) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.PingPong(ctx, &plum.Ping{
		Name: fmt.Sprint("Ping"),
	})

	if err != nil {
		log.Printf("could not ping-pong: %v\n", err)
	}
	log.Println(r.GetMessage())

}

func calculateTps(client plum.FarmerClient, _ time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timer := time.NewTimer(time.Second * 10)

	stream, err := client.GetPeerStateStream(ctx, &plum.Empty{})
	if err != nil {
		log.Fatalf("could not get peer state by streaming: %v", err)
	}

	rwMutex := &sync.RWMutex{}
	blockHeights := make([]uint64, 0, 20)

	go func() {
		for {
			select {
			case <-timer.C:
				return
			default:
				ps, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err == context.Canceled {
					log.Println("context has expired")
				}

				if err != nil {
					log.Fatalf("colud not get peer state by streaming: %v", err)
				}

				rwMutex.Lock()
				//log.Println(ps.BlockHeight)
				blockHeights = append(blockHeights, ps.BlockHeight)
				rwMutex.Unlock()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var base int
		var ticked int
		for {
			select {
			case <-ticker.C:
				ticked++
				rwMutex.RLock()
				tps := blockHeights[len(blockHeights)-1] - blockHeights[base]
				avgTps := float32(blockHeights[len(blockHeights)-1]-blockHeights[0]) / float32(ticked)
				log.Printf("TPS: %d (avg: %f)\n", tps, avgTps)
				base = len(blockHeights) - 1
				rwMutex.RUnlock()
			case <-timer.C:
				return
			}
		}
	}()

	select {
	case <-timer.C:
		log.Println("TPS done")
		return
	}
}

func handleGetPeerStateStream(client plum.FarmerClient, latency time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	stream, err := client.GetPeerStateStream(ctx, &plum.Empty{})
	if err != nil {
		log.Fatalf("could not get peer state by streaming: %v", err)
	}

	for {
		ps, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("colud not get peer state by streaming: %v", err)
		}
		log.Println(formatPeerState(ps))
	}
}

func handleGetPeerState(client plum.FarmerClient, latency time.Duration) {

	ctx, cancel := context.WithTimeout(context.Background(), latency)
	defer cancel()

	r, err := client.GetPeerState(ctx, &plum.Empty{})
	if err != nil {
		log.Printf("could not get the state of peer: %v", err)
	}
	log.Println(formatPeerState(r))
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

func handleTriggerConsensus(conn *grpc.ClientConn, latency time.Duration, nextBlockDigest []byte, privateKey ed25519.PrivateKey, nextBlock *plum.Block) {
	for i := 0; i < *iterFlag; i++ {
		c := plum.NewConsensusClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), latency)

		consensusMessage := &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTNewRound,
			Round:  *roundFlag + uint64(i),
			Digest: nextBlockDigest,
			PeerId: 0,
		}

		md, err := proto.Marshal(consensusMessage)
		if err != nil {
			log.Printf("could not marshal the message: %v\n", err)
		}

		sig := ed25519.Sign(privateKey, md)

		r, err := c.ServePBFTPhase(ctx, &plum.PBFTRequest{
			Message:   consensusMessage,
			Signature: sig,
			Block:     nextBlock,
		})

		if err != nil {
			log.Fatalf("could not serve: %v", err)
		}

		//handle message from response as it has 'oneof' field
		log.Printf("%s, %s | %d\n", r.GetStatus().String(), makeString(r), i)
		time.Sleep(latency)
		cancel()
	}
}

func loadProfile() []string {
	var p []string
	if *localFlag == true {
		p = append(p, "localhost:50051")
		p = append(p, "localhost:50061")
		p = append(p, "localhost:50071")
		p = append(p, "localhost:50081")
	} else {
		p = append(p, fmt.Sprintf("%s:%s", os.Getenv("ABC_PLUM_MACHINE0_IPV4"), os.Getenv("ABC_PLUM_MACHINE0_PORT")))
		p = append(p, fmt.Sprintf("%s:%s", os.Getenv("ABC_PLUM_MACHINE1_IPV4"), os.Getenv("ABC_PLUM_MACHINE1_PORT")))
		p = append(p, fmt.Sprintf("%s:%s", os.Getenv("ABC_PLUM_MACHINE2_IPV4"), os.Getenv("ABC_PLUM_MACHINE2_PORT")))
		p = append(p, fmt.Sprintf("%s:%s", os.Getenv("ABC_PLUM_MACHINE3_IPV4"), os.Getenv("ABC_PLUM_MACHINE3_PORT")))
	}
	return p
}

func makeString(r *plum.PBFTResponse) string {
	var resultMsg string
	switch resultField := r.Result.(type) {
	case *plum.PBFTResponse_Msg:
		resultMsg = string(resultField.Msg)
	case *plum.PBFTResponse_Error:
		resultMsg = resultField.Error.String()
	}
	return resultMsg
}

func setInitialBlockState() (*plum.Block, []byte) {
	gb := ledger.LoadGenesisBlock(path.GetInstance().GenesisBlockPath)
	gbd := block.Digest(gb.Header)
	nextBlock := block.NewBlock(nil, gbd, 1)
	nextBlockDigest := block.Digest(nextBlock.Header)
	return nextBlock, nextBlockDigest
}

func setLatency() time.Duration {
	var latency time.Duration
	switch *speedFlag {
	case 0:
		latency = time.Millisecond * 1000
		log.Println("speed set to: ", latency)
	case 1:
		latency = time.Millisecond * 500
		log.Println("speed set to: ", latency)
	case 2:
		latency = time.Millisecond * 5
		log.Println("speed set to: ", latency)
	default:
		latency = time.Millisecond * 1000
		log.Println("no such value | speed set to default(0): ", latency)
	}
	return latency
}

func setPrivateKey(err error, privateKey ed25519.PrivateKey) ed25519.PrivateKey {
	seed, decodeErr := hex.DecodeString(os.Getenv(fmt.Sprintf("ABC_PLUM_MACHINE%d_SEED", 0)))
	if decodeErr != nil {
		log.Fatalf("could not get seed from string: %v", err)
	}
	privateKey = ed25519.NewKeyFromSeed(seed)
	return privateKey
}

func makeConnection(conn *grpc.ClientConn, err error, profile []string) (*grpc.ClientConn, error) {
	if *localFlag == true {
		conn, err = grpc.Dial(profile[0], grpc.WithInsecure())
	} else {
		switch *targetFlag {
		case "PLUM0":
			conn, err = grpc.Dial(profile[0], grpc.WithInsecure(), grpc.WithBlock())
		case "PLUM1":
			conn, err = grpc.Dial(profile[1], grpc.WithInsecure(), grpc.WithBlock())
		case "PLUM2":
			conn, err = grpc.Dial(profile[2], grpc.WithInsecure(), grpc.WithBlock())
		case "PLUM3":
			conn, err = grpc.Dial(profile[3], grpc.WithInsecure(), grpc.WithBlock())
		default:
			log.Printf("as there is no target set, send to: %v", profile[0])
			conn, err = grpc.Dial(profile[0], grpc.WithInsecure(), grpc.WithBlock())
		}
	}
	return conn, err
}
