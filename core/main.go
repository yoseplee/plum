package main

import (
	"flag"
	"fmt"
	"github.com/yoseplee/plum/core/peer"
	"github.com/yoseplee/plum/core/util"
	"github.com/yoseplee/plum/core/util/path"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var (
	idFlag            = flag.Int("id", -1, "identification number of this peer")
	localOptFlag      = flag.Bool("local", false, "option for executing in local or distributed environment")
	localPortOpTFlag  = flag.String("lport", ":50051", "option for local port")
	dockerModeFlag    = flag.Bool("docker", true, "option for docker environment")
	peerAmountFlag    = flag.Int("amount", -1, "set amount of peers participating in consensus")
	consensusTypeFlag = flag.String("consensus", "XBFT", "option for consensusType")
)

type profile struct {
	Profile []map[string]interface{} `yaml:",flow"`
}

//loadProfile creates connection profile in the peer
//if amount is -1, then it means that make connection over all the peers in the profile
func loadProfile(pr profile, amount int) map[uint32]*peer.Connection {
	if amount == -1 {
		amount = len(pr.Profile)
	}

	if amount > len(pr.Profile) || amount < -1 {
		log.Fatalf("invalid amount: %v, profile length: %v", amount, len(pr.Profile))
	}

	profile := make(map[uint32]*peer.Connection)

	if *localOptFlag == true && *dockerModeFlag == false {
		log.Printf("profile is set for local-dockerless test mode")
		basePort := 50051
		for i := 0; i < amount; i++ {
			port := fmt.Sprintf(":%d", basePort+(10*i))
			profile[uint32(i)] = &peer.Connection{
				PeerId:        uint32(i),
				ContainerName: "",
				Ipv4:          "localhost",
				Port:          port,
			}
		}
		return profile
	}

	if *localOptFlag == true && *dockerModeFlag == true {
		log.Println("profile is set for local-docker mode")
		port := ":50051"
		for i := 0; i < amount; i++ {
			profile[uint32(i)] = &peer.Connection{
				PeerId:        uint32(i),
				ContainerName: fmt.Sprintf("plum%d", i),
				Ipv4:          "",
				Port:          port,
			}
		}
		return profile
	}

	if *localOptFlag == false && *dockerModeFlag == true {
		log.Printf("profile is set for distributed docker mode")
		for _, p := range pr.Profile {
			for _, v := range p {
				mv := v.(map[interface{}]interface{})
				id := mv["PEER_ID"].(int)
				profile[uint32(id)] = &peer.Connection{
					PeerId:        uint32(id),
					ContainerName: mv["CONTAINER_NAME"].(string),
					Ipv4:          mv["IPV4"].(string),
					Port:          mv["PORT"].(string),
				}
			}
		}
		return profile
	}

	if *localOptFlag == false && *dockerModeFlag == false {
		log.Printf("profile is set for distributed process mode")
		for i, p := range pr.Profile {
			if i == amount {
				break
			}
			for _, v := range p {
				mv := v.(map[interface{}]interface{})
				id := mv["PEER_ID"].(int)
				profile[uint32(id)] = &peer.Connection{
					PeerId:        uint32(id),
					ContainerName: "",
					Ipv4:          mv["IPV4"].(string),
					Port:          mv["PORT"].(string),
				}
			}
		}
		return profile
	}

	log.Fatalln("invalid options to make a new profile")
	return nil
}

func main() {

	flag.Parse()

	profile := profile{}

	yamlFile, readErr := ioutil.ReadFile(path.GetInstance().ProfilePath)
	if readErr != nil {
		log.Fatalf("could not read profile.yaml file: %v", readErr)
	}

	yamlParseErr := yaml.Unmarshal(yamlFile, &profile)
	if yamlParseErr != nil {
		log.Fatalf("could not unmarshal profile.yaml file: %v", yamlParseErr)
	}

	peerInstance := peer.GetInstance()

	ipv4, err := util.GetExternalIP()
	if err != nil {
		log.Fatalf("could not get external ip: %v", err)
	}

	loadedProfile := loadProfile(profile, *peerAmountFlag)

	if *localOptFlag == false {
		ipv4 = loadedProfile[uint32(*idFlag)].Ipv4
	}

	//Init the peer
	peerInstance.InitAndRun(uint32(*idFlag), ipv4, *localPortOpTFlag, loadedProfile, *consensusTypeFlag)

	//logging will be printed via below
	switch *idFlag {
	case -1:
		log.Printf("There is no id for this peer... set default: %d\n", *idFlag)
	default:
		log.Printf("peer id is set to: %d\n", peerInstance.ID)
	}

	log.Println("on ipv4: ", peerInstance.Ipv4)
	log.Println("on port: ", peerInstance.Port)

	//go func() {
	//	for {
	//		peerInstance.PrintPeer()
	//		<-time.After(time.Second * 2)
	//	}
	//}()

	s := peer.NewServer()
	s.RegisterServers()
	s.Run()
}
