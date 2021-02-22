package peer

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/yoseplee/plum/core/plum"
	"github.com/yoseplee/plum/core/util"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type server struct {
	plum.UnimplementedConsensusServer
	plum.UnimplementedGossipServer
	plum.UnimplementedFarmerServer
	plum.UnimplementedPeerServer
	gs  *grpc.Server
	lis net.Listener
}

func NewServer() *server {
	return &server{}
}

func (s *server) RegisterServers() {
	s.setNewGrpcServer()
	plum.RegisterConsensusServer(s.gs, s)
	plum.RegisterFarmerServer(s.gs, s)
	plum.RegisterPeerServer(s.gs, s)
}

func (s *server) setNewGrpcServer() {
	gs := grpc.NewServer()
	s.gs = gs
}

func (s *server) Run() {
	p := GetInstance()
	port := p.AddressBook[p.ID].Port

	s.NewListener(port)

	if err := s.gs.Serve(s.lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) NewListener(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s.lis = lis
}

func (s *server) Stop() {
	s.gs.Stop()
	err := s.lis.Close()
	if err != nil {
		log.Printf("could not close listner: %v", err)
	}
}

func (s *server) ServePBFTPhase(_ context.Context, in *plum.PBFTRequest) (*plum.PBFTResponse, error) {
	//log.Println("GOT: ", util.MakeString(in))
	p := GetInstance()

	switch in.Message.GetPhase() {
	case plum.PBFTPhase_PBFTNewRound:
		p.MQ.Push(in)
	case plum.PBFTPhase_PBFTPrePrepare:
		p.MQ.Push(in)
	case plum.PBFTPhase_PBFTPrepare:
		p.MQ.Push(in)
	case plum.PBFTPhase_PBFTCommit:
		p.MQ.Push(in)
	case plum.PBFTPhase_PBFTRoundChange:
		p.MQ.Push(in)
	default:
		//invalid phase
		return &plum.PBFTResponse{
			Status: plum.ResponseStatus_Failed,
			Result: &plum.PBFTResponse_Error{Error: plum.ConsensusValidationCode_Invalid},
		}, nil
	}

	return &plum.PBFTResponse{
		Status: plum.ResponseStatus_Success,
		Result: &plum.PBFTResponse_Msg{Msg: []byte(fmt.Sprintf("%s at %s%s", "the message has received by the peer", p.Ipv4, p.Port))},
	}, nil
}

func (s *server) ServeXBFTPhase(_ context.Context, in *plum.XBFTRequest) (*plum.XBFTResponse, error) {
	//log.Println("GOT: ", util.MakeString(in))
	p := GetInstance()

	switch in.Message.GetPhase() {
	case plum.XBFTPhase_XBFTSelect:
		p.XBFTMQ.Push(in)
	case plum.XBFTPhase_XBFTNewRound:
		p.XBFTMQ.Push(in)
	case plum.XBFTPhase_XBFTPrePrepare:
		p.XBFTMQ.Push(in)
	case plum.XBFTPhase_XBFTPrepare:
		p.XBFTMQ.Push(in)
	case plum.XBFTPhase_XBFTCommit:
		p.XBFTMQ.Push(in)
	case plum.XBFTPhase_XBFTRoundChange:
		p.XBFTMQ.Push(in)
	default:
		//invalid phase
		return &plum.XBFTResponse{
			Status: plum.ResponseStatus_Failed,
			Result: &plum.XBFTResponse_Error{Error: plum.ConsensusValidationCode_Invalid},
		}, nil
	}

	return &plum.XBFTResponse{
		Status: plum.ResponseStatus_Success,
		Result: &plum.XBFTResponse_Msg{Msg: []byte(fmt.Sprintf("%s at %s%s", "the message has received by the peer", p.Ipv4, p.Port))},
	}, nil
}

func (s *server) GetPeerState(_ context.Context, _ *plum.Empty) (*plum.PeerState, error) {
	return getPeerState()
}

func getPeerState() (*plum.PeerState, error) {
	p := GetInstance()
	//if p is not initiated yet, this is an error
	if p.K == nil || p.D == nil {
		return nil, errors.New("the peer is not initiated yet")
	}

	switch p.D.ConsensusType {
	case "PBFT":
		//convert
		m := make(map[int32]int32)
		m[0] = int32(p.PBFTVote[plum.PBFTPhase_PBFTRoundChange])
		m[1] = int32(p.PBFTVote[plum.PBFTPhase_PBFTNewRound])
		m[2] = int32(p.PBFTVote[plum.PBFTPhase_PBFTPrePrepare])
		m[3] = int32(p.PBFTVote[plum.PBFTPhase_PBFTPrepare])
		m[4] = int32(p.PBFTVote[plum.PBFTPhase_PBFTCommit])

		ps := &plum.PeerState{
			Id:             p.ID,
			Ipv4:           p.Ipv4,
			Port:           p.Port,
			Role:           p.Role,
			ConsensusRound: p.ConsensusRound,
			CurrentPrimary: p.Primary,
			ConsensusPhase: p.PBFTPhase,
			Vote:           m,
			ConsensusState: p.ConsensusState,
			BlockHeight:    p.L.Height,
			QueueLength:    p.MQ.GetN(),
			HeapLength:     int64(p.D.ReservedPBFTMessage.GetLast()),
		}
		return ps, nil
	case "XBFT":
		//convert
		ps := &plum.PeerState{
			Id:                     p.ID,
			Ipv4:                   p.Ipv4,
			Port:                   p.Port,
			Role:                   p.Role,
			ConsensusRound:         p.ConsensusRound,
			CurrentPrimary:         p.Primary,
			ConsensusPhase:         p.PBFTPhase,
			ConsensusState:         p.ConsensusState,
			BlockHeight:            p.L.Height,
			QueueLength:            p.MQ.GetN(),
			HeapLength:             int64(p.D.ReservedPBFTMessage.GetLast()),
			Reputation:             p.ReputationBook[p.ID],
			SelectedCount:          p.SelectedCount,
			TentativeSelectedCount: p.TentativeSelectedCount,
		}
		return ps, nil
	default:
		return nil, errors.New("invalid consensus type of peer")
	}
}

func (s *server) GetPeerStateStream(_ *plum.Empty, stream plum.Farmer_GetPeerStateStreamServer) error {

	t := time.NewTimer(time.Second * 500)
	errSig := make(chan error, 1)
	go func() {
		for {
			ps, err := getPeerState()
			if err != nil {
				log.Println("could not get state of peer:", err)
			}
			if err := stream.Send(ps); err != nil {
				errSig <- err
			}
			<-time.After(time.Millisecond * 100)
		}
	}()

	select {
	case <-t.C:
		log.Println("timed out, let the stream be closed")
	case err := <-errSig:
		return err
	}
	return nil
}

func (s *server) SetPublicKey(_ context.Context, pub *plum.PublicKey) (*plum.Empty, error) {
	//verify public key and related information
	//then update the public key in the address book
	log.Printf("pub key get from [ %d ]: %s", pub.Id, hex.EncodeToString(pub.Key))
	instance.AddressBook[pub.GetId()].PublicKey = pub.Key
	return &plum.Empty{}, nil
}

func (s *server) GetPublicKey(_ context.Context, _ *plum.Empty) (*plum.PublicKey, error) {

	//if the peer hasn't initiated yet,
	if instance.D == nil || instance.K == nil {
		return nil, errors.New("the peer hasn't initiated yet")
	}

	return &plum.PublicKey{
		Id:   instance.ID,
		Ipv4: instance.Ipv4,
		Port: instance.Port,
		Key:  instance.PublicKey,
	}, nil
}

func (s *server) GetPublicKeyAllStream(_ *plum.Empty, stream plum.Peer_GetPublicKeyAllStreamServer) error {
	//send all the address book
	for _, address := range instance.AddressBook {
		pk := &plum.PublicKey{
			Id:   address.PeerId,
			Ipv4: address.Ipv4,
			Port: address.Port,
			Key:  address.PublicKey,
		}

		if err := stream.Send(pk); err != nil {
			log.Fatalf("could not send addresses: %v", err)
		}
	}
	return nil
}

func (s *server) PingPong(ctx context.Context, in *plum.Ping) (*plum.Pong, error) {
	log.Printf("Ping from - %v", in.GetName())
	address, err := util.GetExternalIP()
	if err != nil {
		log.Fatalf("failed to get external ip address: %v", err)
	}

	message := fmt.Sprintf("Pong - from %s", address)

	return &plum.Pong{
		Message: message,
	}, nil
}
