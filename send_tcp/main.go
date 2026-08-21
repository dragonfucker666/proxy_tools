package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: send-tcp <host> <port>")
		os.Exit(1)
	}

	hostArg := os.Args[1]
	port := os.Args[2]
	host := stripHostPrefix(hostArg)

	socketPath := os.Getenv("INPUT")
	if socketPath == "" {
		fmt.Println("Error: please set the INPUT environment variable to a socket path")
		os.Exit(1)
	}

	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error: could not listen on socket:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening on", socketPath)
	fmt.Println("Sending all connections to", net.JoinHostPort(host, port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, host, port)
	}
}

func stripHostPrefix(hostArg string) string {
	prefixes := []string{"domain:", "ipv4:", "ipv6:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(hostArg, prefix) {
			return strings.TrimPrefix(hostArg, prefix)
		}
	}
	return hostArg
}

func handleConnection(incoming net.Conn, host string, port string) {
	defer incoming.Close()

	address := net.JoinHostPort(host, port)
	outgoing, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to", address, ":", err)
		return
	}
	defer outgoing.Close()

	done := make(chan bool, 2)

	go copyData(outgoing, incoming, done)
	go copyData(incoming, outgoing, done)

	<-done
	<-done
}

func copyData(destination net.Conn, source net.Conn, done chan bool) {
	io.Copy(destination, source)
	done <- true
}
