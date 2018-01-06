package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	pb "github.com/mellowdrifter/bird_info/proto/as_resolve"
	"google.golang.org/grpc"
)

func main() {
	address := flag.String("server", "", "AS resolver address")
	as := flag.Uint("as", 15169, "As AS number to resolve")

	flag.Parse()

	server := fmt.Sprintf("%v:1179", *address)
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	client := pb.NewAsresolverClient(conn)

	resp, err := client.GetAsName(context.Background(), &pb.AsRequest{
		AsNumber: uint32(*as),
	})

	if err != nil {
		log.Fatalf("Failed to get AS name. Error: %v", err)
	}
	fmt.Printf("The AS Name is %v\n", resp.GetAsName())

}
