package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
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
	fmt.Println("send")
	_, err = conn.Write(query)
	if err != nil {
		fmt.Println("Unable to send query")
	}

	// read response
	fmt.Println("reading")
	n, err = conn.Read(buf[:])
	if err != nil {
		fmt.Println("Error on reading")
	}

	output := string(buf[:n])
	fmt.Printf(output)
}
