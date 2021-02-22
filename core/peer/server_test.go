package peer

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"github.com/yoseplee/plum/core/plum"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func connectClientForServerTest() {
	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not dial to grpc server: %v", err)
	}
	conn = connection
}

func TestServer_SetPublicKey(t *testing.T) {
	//scenario: a client set public key to the peer, on id 4
	//then check if there public key is set on the peer
	pc := plum.NewPeerClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pub, _, kErr := ed25519.GenerateKey(nil)
	if kErr != nil {
		t.Errorf("could not make key: %v", kErr)
	}

	_, err := pc.SetPublicKey(ctx, &plum.PublicKey{
		Id:   4,
		Ipv4: "localhost",
		Port: "50051",
		Key:  pub,
	})

	if err != nil {
		t.Errorf("could not set public key properly: %v", err)
	}

	//compare
	if bytes.Compare(instance.AddressBook[4].PublicKey, pub) != 0 {
		t.Errorf("invalid key comparison between the two key generated and set")
	}
}

func TestServer_GetPublicKey(t *testing.T) {

	pc := plum.NewPeerClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, err := pc.GetPublicKey(ctx, &plum.Empty{})
	if err != nil {
		t.Errorf("could not get public key from the peer: %v", err)
	}

	if bytes.Compare(r.Key, instance.PublicKey) != 0 {
		t.Errorf("got invalid key. got: %s, want: %s", hex.EncodeToString(r.Key), hex.EncodeToString(instance.PublicKey))
	}
}
