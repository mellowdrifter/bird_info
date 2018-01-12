package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:1179", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to start gRPC connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBirdCommClient(conn)

	resp, err := client.CheckConfig(context.Background(), &pb.Family{
		Af: 4,
	})
	if err != nil {
		log.Fatalf("Received an error from gRPC server: %v", err)
	}
	fmt.Printf("%+v\n", resp)

	resp, err = client.ReloadConfig(context.Background(), &pb.Family{
		Af: 6,
	})
	if err != nil {
		log.Fatalf("Received an error from gRPC server: %v", err)
	}
	fmt.Printf("%+v\n", resp)
}
