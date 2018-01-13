package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"text/template"

	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

type neighbour struct {
	Routerid string
	Name     string
	LocalAS  uint32
	Address  string
	AS       uint32
	Password string
}

func main() {
	type peers []neighbour
	var myPeers = peers{
		{
			Routerid: "192.168.1.1",
			Name:     "peer1",
			LocalAS:  12345,
			Address:  "100.100.100.100",
			AS:       100,
		},
		{
			Routerid: "192.168.1.1",
			Name:     "peer2",
			LocalAS:  12345,
			Address:  "200.200.200.200",
			AS:       200,
			Password: "password123",
		},
	}

	t := template.Must(template.New("config").Parse(config))

	t.Execute(os.Stdout, myPeers)

	//err := readConfig()
	log.Println("Listening on port 1179")
	lis, err := net.Listen("tcp", ":1179")
	if err != nil {
		log.Fatalf("Failed to bind: %v,", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBirdCommServer(grpcServer, &server{})

	grpcServer.Serve(lis)
}

func connectBird(af uint32, command []byte) ([]string, error) {

	var bird string
	switch af {
	case 4:
		bird = "/var/run/bird/bird.ctl"
	case 6:
		bird = "/var/run/bird/bird6.ctl"
	default:
		return []string{}, fmt.Errorf("Need to pass a supported address family")
	}
	conn, err := net.Dial("unix", bird)
	if err != nil {
		return []string{}, err
	}
	defer conn.Close()

	// read welcome message
	buf := make([]byte, 4096)
	n, err := conn.Read(buf[:])
	if err != nil {
		return []string{}, err
	}

	// send request
	_, err = conn.Write(command)
	if err != nil {
		return []string{}, err
	}

	// read response
	n, err = conn.Read(buf[:])
	if err != nil {
		return []string{}, err
	}

	output := string(buf[:n])
	return strings.Split(output, "\n"), nil

}

func (s *server) ReloadConfig(ctx context.Context, f *pb.Family) (*pb.ConfReply, error) {
	log.Printf("Received request to ReloadConfig with argument %v", f.GetAf())
	query := []byte("configure\n")

	reply, err := connectBird(f.GetAf(), query)
	if err != nil {
		return &pb.ConfReply{}, err
	}

	for _, line := range reply {
		if strings.Contains(line, "Reconfigured") {
			return &pb.ConfReply{
				Reply:   line,
				Success: true,
			}, nil
		}
	}
	// Return empty string and false if config check not ok
	return &pb.ConfReply{}, fmt.Errorf("Error on reloading")
}

func (s *server) CheckConfig(ctx context.Context, f *pb.Family) (*pb.ConfReply, error) {
	log.Printf("Received request to CheckConfig with argument %v", f.GetAf())
	query := []byte("configure check\n")

	reply, err := connectBird(f.GetAf(), query)
	if err != nil {
		return &pb.ConfReply{}, err
	}

	for _, line := range reply {
		if strings.Contains(line, "Configuration OK") {
			return &pb.ConfReply{
				Reply:   line,
				Success: true,
			}, nil
		}
	}
	// Return empty string and false if config check not ok
	return &pb.ConfReply{}, fmt.Errorf("Error with checking config")
}

func readConfig() error {
	existing, err := ioutil.ReadFile("/etc/bird/bird6.conf")
	if err != nil {
		return fmt.Errorf("Error reading file")
	}
	err = ioutil.WriteFile("/etc/bird/bird5.conf", existing, 0600)
	if err != nil {
		return fmt.Errorf("Error writing file")
	}
	fmt.Println(string(existing))
	return nil
}
