package main

import (
	"flag"
	"fmt"
	"log"

	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {

	// Get new peer data
	server := flag.String("server", "localhost", "Config server")
	name := flag.String("name", "", "new peer name")
	desc := flag.String("desc", "", "description of peer")
	action := flag.String("action", "", "action to perform")
	address := flag.String("address", "", "address of peer")
	as := flag.Uint("as", 0, "as number of peer")
	prefix := flag.String("prefix", "", "A prefix to get to")
	mask := flag.Uint("mask", 0, "a subnet mask")
	nexthop := flag.String("nexthop", "", "A nexthop")

	flag.Parse()

	if *action == "" {
		log.Fatalf("Need an action to perform")
	}

	// Set up connection to server
	serverConn := fmt.Sprintf("%v:1179", *server)
	conn, err := grpc.Dial(serverConn, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect to server: %v", err)
	}
	defer conn.Close()
	client := pb.NewBirdCommClient(conn)

	// Fill peer and route messages with parsed data
	peer := pb.Peer{
		Address:     *address,
		Name:        *name,
		Description: *desc,
		As:          uint32(*as),
	}

	route := pb.Route{
		Prefix:  *prefix,
		Mask:    uint32(*mask),
		Nexthop: *nexthop,
	}

	// check action
	switch *action {
	case "AddPeer":
		res, err := peerAction(&peer, client, true)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	case "DeletePeer":
		res, err := peerAction(&peer, client, false)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	case "AddRoute":
		res, err := routeAction(&route, client, true)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	case "DeleteRoute":
		res, err := routeAction(&route, client, false)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	default:
		log.Fatalf("Must select a supported action (AddPeer, DeletePeer, AddRoute, DeleteRoute)")
	}
}

func peerAction(p *pb.Peer, client pb.BirdCommClient, add bool) (*pb.Result, error) {
	// Add peer if it's new
	if add {
		resp, err := client.AddNeighbour(context.Background(), p)
		if err != nil {
			log.Fatalf("%v", err)
		}
		return resp, err
	}

	// delete peer otherwise
	resp, err := client.DeleteNeighbour(context.Background(), p)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return resp, err
}

func routeAction(r *pb.Route, client pb.BirdCommClient, add bool) (*pb.Result, error) {
	// Add route if it's new
	if add {
		resp, err := client.AddStatic(context.Background(), r)
		if err != nil {
			log.Fatalf("%v", err)
		}
		return resp, err
	}

	// delete route otherwise
	resp, err := client.DeleteStatic(context.Background(), r)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return resp, err
}
