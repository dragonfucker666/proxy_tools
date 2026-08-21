package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func main() {
	inputPath := os.Getenv("INPUT")
	outputPath := os.Getenv("OUTPUT")

	if inputPath == "" || outputPath == "" {
		log.Fatal("Please set INPUT and OUTPUT to unix socket paths")
	}

	os.Remove(inputPath)

	listener, err := net.Listen("unix", inputPath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConnection(clientConn, outputPath)
	}
}

func handleConnection(clientConn net.Conn, outputPath string) {
	defer clientConn.Close()

	backendConn, err := net.Dial("unix", outputPath)
	if err != nil {
		log.Println("dial error:", err)
		return
	}
	defer backendConn.Close()

	clientReader := bufio.NewReader(clientConn)

	err = stripRequestHeaders(clientReader)
	if err != nil {
		log.Println("could not read request headers:", err)
		return
	}

	err = writeResponseHeaders(clientConn)
	if err != nil {
		log.Println("could not write response headers:", err)
		return
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		io.Copy(backendConn, clientReader)
		closeWriteSide(backendConn)
	}()

	go func() {
		defer waitGroup.Done()
		io.Copy(clientConn, backendConn)
		closeWriteSide(clientConn)
	}()

	waitGroup.Wait()
}

func closeWriteSide(conn net.Conn) {
	unixConn, ok := conn.(*net.UnixConn)
	if ok {
		unixConn.CloseWrite()
	}
}

func stripRequestHeaders(reader *bufio.Reader) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if line == "\r\n" || line == "\n" {
			return nil
		}
	}
}

func writeResponseHeaders(conn net.Conn) error {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	return err
}
