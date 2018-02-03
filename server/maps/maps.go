package main

import (
	"fmt"

	pb "github.com/mellowdrifter/bird_info/proto/mapTest"
)

func main() {
	l3Details := pb.Details{
		Description: "Level 3",
		As:          1,
	}
	gDetails := pb.Details{
		Description: "Google",
		As:          15169,
	}
	peers := pb.Peer{}
	peers.Neighbour = make(map[string]*pb.Details)
	peers.Neighbour["10.1.1.1"] = &l3Details
	peers.Neighbour["10.1.1.2"] = &gDetails

	for k, v := range peers.Neighbour {
		fmt.Printf("%s has name %s and AS number %d\n", k, v.GetDescription(), v.GetAs())
	}
}
