package main

import (
	"log"
	"io"
	"net"
	"os"
	"flag"
)

func main() {
	ipv := flag.String("ipv", "any", "version of IP protocol to use")
	flag.Parse()

	positionals := flag.Args()

	if len(positionals) < 2 {
		log.Fatalln("Usage: ./listen_tcp <host> <port>")
	}

	host := positionals[0]
	port := positionals[1]

	socketPath := os.Getenv("OUTPUT")
	if socketPath == "" {
		log.Fatalln("Please set OUTPUT env var to a socket path")
	}

	network, ok := map[string]string {
		"any": "tcp",
		"4": "tcp4",
		"6": "tcp6",
	}[*ipv];
	if !ok {
		log.Fatalln("Incorrect IP version: expected \"4\", \"6\" or \"any\", got", *ipv)
	}

	listener, err := net.Listen(network, net.JoinHostPort(host, port))
	if err != nil {
		log.Fatalln("Cannot listen:", err)
	}
	defer listener.Close()

	log.Println("Listening on", host, network)
	log.Println("Sending connections to", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn, socketPath)
	}
}

func handleConnection(clientConn net.Conn, socketPath string) {
	defer clientConn.Close()

	socketConn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Println("Could not connect to socket:", err)
		return
	}
	defer socketConn.Close()

	go io.Copy(socketConn, clientConn)
	io.Copy(clientConn, socketConn)
}
