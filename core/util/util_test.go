package util

import (
	"github.com/yoseplee/plum/core/plum"
	"testing"
)

var (
	Ipv4 string
	Port string
)

func Test_Init(t *testing.T) {
	Ipv4 = "localhost"
	Port = "50051"
}

func TestMakeTarget(t *testing.T) {
	want := "localhost:50051"
	if got, err := MakeTarget(Ipv4, Port); got != want || err != nil {
		t.Errorf("failed to make full path properly. got: %v, want: %v", got, want)
	}
}

func TestMakeString(t *testing.T) {
	var want string
	want = "[MSG: PBFTRequest] peer(0), round(0), phase(PBFTNewRound)"
	if got := MakeString(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTNewRound,
			Round:  0,
			PeerId: 0,
		},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTRequest] peer(0), round(3), phase(PBFTPrePrepare)"
	if got := MakeString(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTPrePrepare,
			Round:  3,
			PeerId: 0,
		},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTRequest] peer(0), round(5000), phase(PBFTPrepare)"
	if got := MakeString(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTPrepare,
			Round:  5000,
			PeerId: 0,
		},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTRequest] peer(0), round(10000), phase(PBFTCommit)"
	if got := MakeString(&plum.PBFTRequest{
		Message: &plum.PBFTMessage{
			Phase:  plum.PBFTPhase_PBFTCommit,
			Round:  10000,
			PeerId: 0,
		},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTResponse] status(Success), result(the message has received by the peer)"
	if got := MakeString(&plum.PBFTResponse{
		Status: plum.ResponseStatus_Success,
		Result: &plum.PBFTResponse_Msg{Msg: []byte("the message has received by the peer")},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTResponse] status(Failed), result(Invalid)"
	if got := MakeString(&plum.PBFTResponse{
		Status: plum.ResponseStatus_Failed,
		Result: &plum.PBFTResponse_Error{Error: plum.ConsensusValidationCode_Invalid},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTResponse] status(Failed), result(TimeOut)"
	if got := MakeString(&plum.PBFTResponse{
		Status: plum.ResponseStatus_Failed,
		Result: &plum.PBFTResponse_Error{Error: plum.ConsensusValidationCode_TimeOut},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTResponse] status(Failed), result(BadRequest)"
	if got := MakeString(&plum.PBFTResponse{
		Status: plum.ResponseStatus_Failed,
		Result: &plum.PBFTResponse_Error{Error: plum.ConsensusValidationCode_BadRequest},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "[MSG: PBFTResponse] status(Failed), result(BadResponse)"
	if got := MakeString(&plum.PBFTResponse{
		Status: plum.ResponseStatus_Failed,
		Result: &plum.PBFTResponse_Error{Error: plum.ConsensusValidationCode_BadResponse},
	}); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	want = "hello this is string"
	if got := MakeString("hello this is string"); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}

	if got := MakeString([]byte("hello this is string")); got != want {
		t.Errorf("invalid stringify on PBFTRequest message. want: %v, got: %v", want, got)
	}
}
