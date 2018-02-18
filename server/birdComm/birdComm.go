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

	"github.com/mellowdrifter/bird_info/proto/birdComm"

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
	// Set up gRPC server
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
	// This reads the cruft off bird and ensure we get clean data back
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

func getConfig(family string) configFiles {
	switch family {
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

func reMarshal(c *configFiles, m proto.Message) error {
	var typeMarshal string
	switch m.(type) {
	case *birdComm.RouteGroup:
		typeMarshal = c.staticMarshal
	case *birdComm.PeerGroup:
		typeMarshal = c.bgpMarshal
	}
	out, _ := os.Create(typeMarshal)

	err := proto.MarshalText(out, m)
	if err != nil {
		return err
	}
	return nil
}

func (s *server) AddNeighbour(ctx context.Context, p *pb.Peer) (*pb.Result, error) {
	log.Printf("Adding peer %v", p.GetAddress())
	// TO-DO - Validate update:
	// Ensure correct values in place
	// Determine IPv4 vs IPv6
	// Determine IP is valid

	// Load config for address family
	// Fix when above validation is done
	conf := getConfig("ipv4")

	// Get existing peers
	peers, err := loadExistingPeers(&conf)
	if err != nil {
		return nil, err
	}

	// If peer already exists with same settings, then we should just return success
	// If any setting is different, we'll just override the current peer with new settings
	peerAddress := p.GetAddress()
	if _, ok := peers.Peer[peerAddress]; ok {
		if proto.Equal(p, peers.Peer[peerAddress]) {
			return &pb.Result{
				Reply:   "Peer already configured",
				Success: true,
			}, nil
		}
	}

	// Append new peer
	newPeers := pb.PeerGroup{}
	newPeers.Peer = make(map[string]*pb.Peer)
	// Is there a better way of doing a map copy?
	for k, v := range peers.GetPeer() {
		newPeers.Peer[k] = v
	}
	newPeers.Peer[peerAddress] = p

	// Write new BGP peer config
	out, err := os.Create(conf.bgpConfig)
	defer out.Close()
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("bgp").Parse(bgp))
	t.Execute(out, newPeers.GetPeer())

	// Check if new config loads. If not we need to rollback to the old config
	resp, err := reloadConfig(&conf)
	if err != nil {
		out, _ := os.Create(conf.bgpConfig)
		defer out.Close()
		t := template.Must(template.New("bgp").Parse(bgp))
		t.Execute(out, peers.GetPeer())
		return resp, err
	}

	// Else we are good.
	// Remarshal the config locally to use for next time
	err = reMarshal(&conf, &newPeers)
	return resp, err
}
func (s *server) DeleteNeighbour(ctx context.Context, p *pb.Peer) (*pb.Result, error) {
	log.Printf("Removing peer %v", p.GetAddress())

	// TO-DO - Validate update:
	// Ensure correct values in place
	// Determine IPv4 vs IPv6
	// Determine IP is valid

	// Load config for address family
	// Fix when above validation is done
	conf := getConfig("ipv4")

	// Get existing peers
	peers, err := loadExistingPeers(&conf)
	if err != nil {
		return nil, err
	}

	// If peer is not configured, just return success
	peerAddress := p.GetAddress()
	if _, ok := peers.Peer[peerAddress]; !ok {
		return &pb.Result{
			Reply:   "Peer not configured",
			Success: true,
		}, nil
	}

	// Create new peergroup and remove peer from the group
	newPeers := pb.PeerGroup{}
	newPeers.Peer = make(map[string]*pb.Peer)
	// Is there a better way of doing a map copy?
	for k, v := range peers.GetPeer() {
		newPeers.Peer[k] = v
	}
	delete(newPeers.GetPeer(), peerAddress)

	// Write new BGP peer config
	out, err := os.Create(conf.bgpConfig)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("bgp").Parse(bgp))
	t.Execute(out, newPeers.GetPeer())
	out.Close()

	// Check if new config loads. If not we need to rollback to the old config
	resp, err := reloadConfig(&conf)
	if err != nil {
		out, _ := os.Create(conf.bgpConfig)
		defer out.Close()
		t := template.Must(template.New("bgp").Parse(bgp))
		t.Execute(out, peers.GetPeer())
		return resp, err
	}

	// Else we are good.
	// Remarshal the config locally to use for next time
	err = reMarshal(&conf, &newPeers)
	return resp, err

}

// TO-DO - Consolidate the following two functions
/*func remove(pg *pb.PeerGroup, p *pb.Peer) pb.PeerGroup {
	var newPeers pb.PeerGroup
	for _, peer := range pg.GetGroup() {
		if !proto.Equal(p, peer) {
			newPeers.Group = append(newPeers.Group, peer)
		}
	}
	return newPeers
}*/

func removeS(rg *pb.RouteGroup, r *pb.Route) pb.RouteGroup {
	var newRoutes pb.RouteGroup
	for _, route := range rg.GetRoutes() {
		if !proto.Equal(r, route) {
			newRoutes.Routes = append(newRoutes.Routes, route)
		}
	}
	return newRoutes
}
func (s *server) AddStatic(ctx context.Context, r *pb.Route) (*pb.Result, error) {
	// Load config for address family
	conf := getConfig("IPv4")

	// Get existing routes
	routes, err := loadExistingRoutes(&conf)
	if err != nil {
		return nil, err
	}

	for _, route := range routes.GetRoutes() {
		if proto.Equal(route, r) {
			return &pb.Result{
				Reply:   "Route already configured",
				Success: true,
			}, nil
		}
	}

	var newRoutes pb.RouteGroup
	newRoutes.Routes = append(routes.Routes, r)

	// Write new static route config
	out, err := os.Create(conf.staticConfig)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("static").Parse(static))
	t.Execute(out, newRoutes.Routes)
	//t.Execute(os.Stdout, newRoutes.Routes)
	out.Close()

	// Check if new config loads. If not we need to rollback to the old config
	resp, err := reloadConfig(&conf)
	if err != nil {
		out, _ := os.Create(conf.staticConfig)
		defer out.Close()
		t := template.Must(template.New("static").Parse(static))
		t.Execute(out, routes.Routes)
		return resp, errors.New("New config not loaded. Restoring old config")
	}

	// Else we are good.
	// Remarshal the config locally to use for next time
	err = reMarshal(&conf, &newRoutes)
	return resp, err
}
func (s *server) DeleteStatic(ctx context.Context, r *pb.Route) (*pb.Result, error) {
	// Load config for address family
	conf := getConfig("IPv4")

	// Get existing routes
	routes, err := loadExistingRoutes(&conf)
	if err != nil {
		return nil, err
	}

	var newRoutes pb.RouteGroup
	newRoutes = removeS(routes, r)

	if len(newRoutes.Routes) == len(routes.Routes) {
		return &pb.Result{
			Reply:   "Route not configured",
			Success: true,
		}, nil
	}

	// Write new static route config
	out, err := os.Create(conf.staticConfig)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("static").Parse(static))
	t.Execute(out, newRoutes.Routes)
	//t.Execute(os.Stdout, newRoutes.Routes)
	out.Close()

	// Check if new config loads. If not we need to rollback to the old config
	resp, err := reloadConfig(&conf)
	if err != nil {
		out, _ := os.Create(conf.staticConfig)
		defer out.Close()
		t := template.Must(template.New("static").Parse(static))
		t.Execute(out, routes.Routes)
		return resp, errors.New("New config not loaded. Restoring old config")
	}

	// Else we are good.
	// Remarshal the config locally to use for next time
	err = reMarshal(&conf, &newRoutes)
	return resp, err
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
	// TO-DO - Return actual error. Otherwise difficult to troubleshoot what went wrong
	return nil, fmt.Errorf("Error on reloading")
}

func loadExistingPeers(c *configFiles) (*pb.PeerGroup, error) {
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

func loadExistingRoutes(c *configFiles) (*pb.RouteGroup, error) {
	in, err := ioutil.ReadFile(c.staticMarshal)
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
