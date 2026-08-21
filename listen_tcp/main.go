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
		fmt.Println("Usage: listen-tcp <host> <port>")
		os.Exit(1)
	}

	hostArg := os.Args[1]
	port := os.Args[2]

	network, host, ok := parseHost(hostArg)
	if !ok {
		fmt.Println("Host must start with domain:, ipv4:, or ipv6:")
		os.Exit(1)
	}

	socketPath := os.Getenv("OUTPUT")
	if socketPath == "" {
		fmt.Println("Please set OUTPUT to a socket path")
		os.Exit(1)
	}

	address := net.JoinHostPort(host, port)
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("Cannot listen:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening on", address)
	fmt.Println("Sending connections to", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn, socketPath)
	}
}

func parseHost(hostArg string) (network string, host string, ok bool) {
	if strings.HasPrefix(hostArg, "domain:") {
		return "tcp", strings.TrimPrefix(hostArg, "domain:"), true
	}
	if strings.HasPrefix(hostArg, "ipv4:") {
		return "tcp4", strings.TrimPrefix(hostArg, "ipv4:"), true
	}
	if strings.HasPrefix(hostArg, "ipv6:") {
		return "tcp6", strings.TrimPrefix(hostArg, "ipv6:"), true
	}
	return "", "", false
}

func handleConnection(clientConn net.Conn, socketPath string) {
	defer clientConn.Close()

	socketConn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Println("Cannot connect to socket:", err)
		return
	}
	defer socketConn.Close()

	go io.Copy(socketConn, clientConn)
	io.Copy(clientConn, socketConn)
}
