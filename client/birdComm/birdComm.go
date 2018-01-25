package main

import (
	"flag"
	"fmt"
	"log"

	"golang.org/x/net/context"

	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	"google.golang.org/grpc"
)

func main() {

	// TO-DO
	// Option to add/delete static routes
	// Also delete neighbours. Should be able to delete via IP address or name

	// Get new peer data
	server := flag.String("server", "localhost", "Config server")
	name := flag.String("name", "", "New peer name")
	desc := flag.String("description,", "", "description of peer")
	// TO-DO
	// Should use an IP library to ensure entered IP is valid
	address := flag.String("address", "", "address of peer")
	as := flag.Uint("as", 0, "as number of peer")
	flag.Parse()

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

	resp, err := client.DeleteNeighbour(context.Background(), &pb.Peer{
		Name:        *name,
		Description: *desc,
		Address:     *address,
		As:          uint32(*as),
		// TO-DO
		// This needs to be added. Or dynamic via IP library?
		Family: pb.Family_ipv4,
	})
	if err != nil {
		log.Fatalf("Received an error from gRPC server: %v", err)
	}
	fmt.Printf("%+v\n", resp)

}
