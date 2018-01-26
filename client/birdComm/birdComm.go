package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mellowdrifter/bird_info/proto/birdComm"

	"golang.org/x/net/context"

	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	"google.golang.org/grpc"
)

func main() {

	// TO-DO
	// Option to add/delete static routes

	// Get new peer data
	server := flag.String("server", "localhost", "Config server")
	name := flag.String("name", "", "new peer name")
	desc := flag.String("description,", "", "description of peer")
	action := flag.String("action", "", "action to perform")
	// TO-DO
	// Should use an IP library to ensure entered IP is valid
	address := flag.String("address", "", "address of peer")
	as := flag.Uint("as", 0, "as number of peer")
	flag.Parse()

	if *action == "" {
		log.Fatalf("Need an action to perform")
	}

	// We need a certain minimum of data to add a neighbour
	if *name == "" || *address == "" || *as == 0 {
		log.Fatalln("At minimum, address, as, and name is required")
	}

	// TO-DO
	// Need to be able to specify address family.... Should also ensure
	// it's valid

	serverConn := fmt.Sprintf("%v:1179", *server)
	conn, err := grpc.Dial(serverConn, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBirdCommClient(conn)
	peer := &pb.Peer{
		Name:        *name,
		Description: *desc,
		Address:     *address,
		As:          uint32(*as),
		Family:      pb.Family_ipv4,
	}

	switch *action {
	case "DeleteNeighbour":
		res, err := delete(peer, client)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	case "AddNeighbour":
		res, err := add(peer, client)
		if err != nil {
			log.Fatalf("error received: %v\n", err)
		}
		fmt.Printf("%v\n", res)
	default:
		log.Fatalf("Action not implemented")
	}
}

func delete(p *pb.Peer, client birdComm.BirdCommClient) (*pb.Result, error) {
	resp, err := client.DeleteNeighbour(context.Background(), p)
	if err != nil {
		log.Fatalf("Received an error from gRPC server: %v", err)
	}
	return resp, err
}

func add(p *pb.Peer, client birdComm.BirdCommClient) (*pb.Result, error) {
	resp, err := client.AddNeighbour(context.Background(), p)
	if err != nil {
		log.Fatalf("Received an error from gRPC server: %v", err)
	}
	return resp, err
}
