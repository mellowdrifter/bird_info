package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type confReply struct {
	reply   string
	success bool
}

type birdConn struct {
	socket string
}

func main() {
	bird4 := birdConn{
		socket: "/var/run/bird/bird.ctl",
	}
	bird6 := birdConn{
		socket: "/var/run/bird/bird6.ctl",
	}

	goCheck := checkConfig()
	if goCheck.success == true {
		fmt.Println("Config check was a success!")
		fmt.Printf("The line we were looking for was %v\n", goCheck.reply)
	} else {
		fmt.Println("Config check was not a success")
	}
}

func checkConfig() confReply {
	query := []byte("configure check\n")
	conn, err := net.Dial("unix", "/var/run/bird/bird6.ctl")
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	// read welcome message
	buf := make([]byte, 4096)
	n, err := conn.Read(buf[:])
	if err != nil {
		fmt.Println("Error on reading")
	}

	// send request
	_, err = conn.Write(query)
	if err != nil {
		fmt.Println("Unable to send query")
	}

	// read response
	n, err = conn.Read(buf[:])
	if err != nil {
		fmt.Println("Error on reading")
	}

	output := string(buf[:n])
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Configuration OK") {
			return confReply{
				reply:   line,
				success: true,
			}
		}
	}
	// Return empty string and false if config check not ok
	return confReply{}
}
