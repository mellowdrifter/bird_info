package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	pb "github.com/mellowdrifter/bird_info/proto/birdComm"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}
type configFiles struct {
	bird          string
	family        uint8
	bgpConfig     string
	staticConfig  string
	bgpMarshal    string
	staticMarshal string
}

func main() {

	log.Println("Listening on port 1179")
	lis, err := net.Listen("tcp", ":1179")
	if err != nil {
		log.Fatalf("Failed to bind: %v,", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBirdCommServer(grpcServer, &server{})

	grpcServer.Serve(lis)
}

func connectBird(c *configFiles, command []byte) ([]string, error) {

	conn, err := net.Dial("unix", c.bird)
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

	// return the rest of the output
	output := string(buf[:n])
	return strings.Split(output, "\n"), nil

}

func getConfig(p *pb.Peer) configFiles {
	switch p.GetFamily().String() {
	case "ipv4":
		return configFiles{
			bird:          "/var/run/bird/bird.ctl",
			family:        4,
			bgpConfig:     "/etc/bird/bird4_bgp.conf",
			staticConfig:  "/etc/bird/bird4_static.conf",
			bgpMarshal:    "neighbours4.pb.txt",
			staticMarshal: "static4.pb.txt",
		}
	case "ipv6":
		return configFiles{
			bird:          "/var/run/bird/bird6.ctl",
			family:        6,
			bgpConfig:     "/etc/bird/bird6_bgp.conf",
			staticConfig:  "/etc/bird/bird6_static.conf",
			bgpMarshal:    "neighbours6.pb.txt",
			staticMarshal: "static6.pb.txt",
		}
	default:
		return configFiles{}
	}
}

func reMarshal(c *configFiles, m *proto.Message) error {
	out, _ := os.Create(c.bgpMarshal)

	err := proto.MarshalText(out, *m)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) AddNeighbour(ctx context.Context, p *pb.Peer) (*pb.Result, error) {

	// Load config for address family
	conf := getConfig(p)

	// Get existing peers
	peers, err := loadExistingPeers(p, &conf)
	if err != nil {
		return nil, err
	}

	// TO-DO
	// If peer already exists with same settings, then we should just return success
	// If any part is different, we should delete that peer and re-add it with new config
	for _, peer := range peers.GetGroup() {
		if proto.Equal(peer, p) {
			return &pb.Result{
				Reply:   "Peer already configured",
				Success: true,
			}, nil
		}
	}

	var newPeers pb.PeerGroup
	// Append new peer
	newPeers.Group = append(peers.Group, p)

	// Write new BGP peer config
	out, err := os.Create(conf.bgpConfig)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("bgp").Parse(bgp))
	t.Execute(out, newPeers.Group)
	out.Close()

	// Check if new config loads. If not we need to rollback to the old config
	resp, err := reloadConfig(&conf)
	if err != nil {
		out, _ := os.Create(conf.bgpConfig)
		defer out.Close()
		t := template.Must(template.New("bgp").Parse(bgp))
		t.Execute(out, peers.Group)
		return resp, errors.New("New config not loaded. Restoring old config")
	}

	// Else we are good.
	// TO-DO - need to marshal the new peer like existing. But where?
	err = reMarshal(&conf, &newPeers)
	return resp, err
}
func (s *server) DeleteNeighbour(ctx context.Context, p *pb.Peer) (*pb.Result, error) {
	return nil, nil
}
func (s *server) AddStatic(ctx context.Context, p *pb.Route) (*pb.Result, error) {
	return nil, nil
}
func (s *server) DeleteStatic(ctx context.Context, p *pb.Route) (*pb.Result, error) {
	return nil, nil
}

func reloadConfig(c *configFiles) (*pb.Result, error) {
	query := []byte("configure\n")

	reply, err := connectBird(c, query)
	if err != nil {
		return nil, err
	}

	for _, line := range reply {
		if strings.Contains(line, "Reconfigured") || strings.Contains(line, "Reconfiguration in progress") {
			return &pb.Result{
				Reply:   line,
				Success: true,
			}, nil
		}
	}
	// Return empty string and false if config check not ok
	return nil, fmt.Errorf("Error on reloading")
}

func loadExistingPeers(p *pb.Peer, c *configFiles) (*pb.PeerGroup, error) {
	in, err := ioutil.ReadFile(c.bgpMarshal)
	if err != nil {
		return nil, err
	}

	peers := &pb.PeerGroup{}
	err = proto.UnmarshalText(string(in), peers)
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func loadExistingRoutes() (*pb.RouteGroup, error) {
	in, err := ioutil.ReadFile("routes.pb.txt")
	if err != nil {
		return nil, err
	}

	routes := &pb.RouteGroup{}
	err = proto.UnmarshalText(string(in), routes)
	if err != nil {
		return nil, err
	}
	return routes, nil
}
