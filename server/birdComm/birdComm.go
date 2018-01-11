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

func main() {

	goCheck := checkConfig()
	if goCheck.success == true {
		fmt.Println("Config check was a success!")
		fmt.Printf("The line we were looking for was %v\n", goCheck.reply)
	} else {
		fmt.Println("Config check was not a success")
	}

	goReload := reloadConfig()
	if goReload.success == true {
		fmt.Println("Config check was a success!")
		fmt.Printf("The line we were looking for was %v\n", goReload.reply)
	} else {
		fmt.Println("Config reload was not a success")
	}

}

func connectBird(af int, command []byte) ([]string, error) {

	var bird string
	switch af {
	case 4:
		bird = "/var/run/bird/bird.ctl"
	case 6:
		bird = "/var/run/bird/bird6.ctl"
	default:
		return []string{}, fmt.Errorf("Need to pass an address family")
	}
	conn, err := net.Dial("unix", bird)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	// read welcome message
	buf := make([]byte, 4096)
	n, err := conn.Read(buf[:])
	if err != nil {
		return []string{}, fmt.Errorf("Error on reading")
	}

	// send request
	_, err = conn.Write(command)
	if err != nil {
		return []string{}, fmt.Errorf("Unable to send query")
	}

	// read response
	n, err = conn.Read(buf[:])
	if err != nil {
		return []string{}, fmt.Errorf("Error on reading")
	}

	output := string(buf[:n])
	return strings.Split(output, "\n"), nil

}

func reloadConfig() confReply {
	query := []byte("configure\n")

	reply, err := connectBird(4, query)
	if err != nil {
		log.Fatalf("received error: %v", err)
		return confReply{}
	}

	for _, line := range reply {
		if strings.Contains(line, "Reconfigured") {
			return confReply{
				reply:   line,
				success: true,
			}
		}
	}
	// Return empty string and false if config check not ok
	return confReply{}
}

func checkConfig() confReply {
	query := []byte("configure check\n")

	reply, err := connectBird(4, query)
	if err != nil {
		log.Fatalf("received error: %v", err)
		return confReply{}
	}

	for _, line := range reply {
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
