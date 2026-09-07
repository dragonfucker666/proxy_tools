package main

import (
	"log"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"flag"
)

func main() {
	ipv := flag.String("ipv", "any", "version of IP protocol to use")
	flag.Parse()

	positionals := flag.Args()

	if len(positionals) < 2 {
		log.Fatalln("Usage: ./send_tcp <host> <port>")
	}

	host := positionals[0]
	port := positionals[1]

	socketPath := os.Getenv("INPUT")
	if socketPath == "" {
		log.Fatalln("Please set INPUT env var to a socket path")
	}

	network, ok := map[string]string {
		"any": "tcp",
		"4": "tcp4",
		"6": "tcp6",
	}[*ipv];
	if !ok {
		log.Fatalln("Incorrect IP version: expected \"4\", \"6\" or \"any\", got", *ipv)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalln("Cannot listen:", err)
	}
	defer listener.Close()

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go func() {
			defer client.Close()
			server, err := net.Dial(network, net.JoinHostPort(host, port))
			if err != nil {
				log.Println("Could not connect to server:", err)
			}
			defer server.Close()
			go io.Copy(server, client)
			io.Copy(client, server)
		}()
	}
}
