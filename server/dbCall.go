package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "github.com/go-sql-driver/mysql"
	pb "github.com/mellowdrifter/bird_info/proto/as_resolve"
	"github.com/spf13/viper"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Struct to hold the gRPC server
type asServer struct{}

// Database details
type dbInfo struct {
	db, user, pass, host, port string
}

var connection dbInfo

func main() {
	// get database details
	connection = readConfig()

	//main will set up the gRPC server and listen for incoming requests
	fmt.Println("Listening on port 1179")
	lis, err := net.Listen("tcp", ":1179")
	if err != nil {
		log.Fatalf("failed to bind to port: %v", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pb.RegisterAsresolverServer(grpcServer, &asServer{})
	grpcServer.Serve(lis)

}

func readConfig() dbInfo {
	viper.SetConfigName("db_config")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	config := dbInfo{
		db:   viper.GetString("db"),
		user: viper.GetString("user"),
		pass: viper.GetString("pass"),
		host: viper.GetString("host"),
		port: viper.GetString("port"),
	}
	return config
}

func resolve(asNumber uint32) (string, error) {
	/* resolve will connect to the database configured and pull the
	autonomous system name from the number given */

	var name string // to hold the AS name
	conn := fmt.Sprintf(`%v:%v@tcp(%v:%v)/%v`,
		connection.user,
		connection.pass,
		connection.host,
		connection.port,
		connection.db)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return "", err
	}
	err = db.QueryRow("select AS_NAME from ASN where AS_NUM = ?", asNumber).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (s *asServer) GetAsName(ctx context.Context, req *pb.AsRequest) (*pb.AsResponse, error) {
	/* GetAsName is the RPC function that will take in an AS number and resolve it to
	an autonamous system name */
	number := req.GetAsNumber()
	fmt.Printf("Received the AS number %v\n", number)

	asName, err := resolve(number)
	if err != nil {
		return &pb.AsResponse{
			AsName: "",
		}, err
	}
	return &pb.AsResponse{
		AsName: asName,
	}, nil

}
