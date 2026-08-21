package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"sync"
)

const frameReal = 0
const frameTrash = 1

func writeFrame(conn net.Conn, frameType byte, data []byte) error {
	header := []byte{frameType, byte(len(data))}
	if _, err := conn.Write(header); err != nil {
		return err
	}
	if len(data) > 0 {
		if _, err := conn.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func readFrame(conn net.Conn) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, err
	}
	frameType := header[0]
	length := header[1]
	data := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(conn, data); err != nil {
			return 0, nil, err
		}
	}
	return frameType, data, nil
}

func trashAmount(realLength int, ratio float64) int {
	target := float64(realLength) * ratio
	amount := int(target)
	leftover := target - float64(amount)
	if rand.Float64() < leftover {
		amount++
	}
	if amount > 255 {
		amount = 255
	}
	return amount
}

func randomBytes(length int) []byte {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte(rand.IntN(256))
	}
	return data
}

func trashCopy(src net.Conn, dst net.Conn, ratio float64) {
	readBuf := make([]byte, 200)
	for {
		n, readErr := src.Read(readBuf)
		if n > 0 {
			realBytes := readBuf[:n]
			splitPoint := rand.IntN(n + 1)
			trashBytes := randomBytes(trashAmount(n, ratio))

			if err := writeFrame(dst, frameReal, realBytes[:splitPoint]); err != nil {
				return
			}
			if err := writeFrame(dst, frameTrash, trashBytes); err != nil {
				return
			}
			if err := writeFrame(dst, frameReal, realBytes[splitPoint:]); err != nil {
				return
			}
		}
		if readErr != nil {
			return
		}
	}
}

func untrashCopy(src net.Conn, dst net.Conn) {
	for {
		frameType, data, err := readFrame(src)
		if err != nil {
			return
		}
		if frameType == frameReal && len(data) > 0 {
			if _, err := dst.Write(data); err != nil {
				return
			}
		}
	}
}

func handleConnection(inputConn net.Conn, outputPath string, ratio float64, clean bool) {
	outputConn, err := net.Dial("unix", outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not connect to OUTPUT socket:", err)
		inputConn.Close()
		return
	}

	inputUnix := inputConn.(*net.UnixConn)
	outputUnix := outputConn.(*net.UnixConn)

	var directionsDone sync.WaitGroup
	directionsDone.Add(2)

	go func() {
		defer directionsDone.Done()
		if clean {
			untrashCopy(inputConn, outputConn)
		} else {
			trashCopy(inputConn, outputConn, ratio)
		}
		outputUnix.CloseWrite()
	}()

	go func() {
		defer directionsDone.Done()
		if clean {
			trashCopy(outputConn, inputConn, ratio)
		} else {
			untrashCopy(outputConn, inputConn)
		}
		inputUnix.CloseWrite()
	}()

	directionsDone.Wait()
	inputConn.Close()
	outputConn.Close()
}

func main() {
	ratio := flag.Float64("trash-ratio", 0.3, "how much trash to add, for example 0.3 means 30 percent")
	clean := flag.Bool("clean", false, "clean INPUT to OUTPUT instead of trashing it, and trash OUTPUT to INPUT instead")
	flag.Parse()

	inputPath := os.Getenv("INPUT")
	outputPath := os.Getenv("OUTPUT")

	if inputPath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "please set INPUT and OUTPUT to unix socket paths")
		os.Exit(1)
	}

	os.Remove(inputPath)
	listener, err := net.Listen("unix", inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not listen on INPUT socket:", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		inputConn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "accept error:", err)
			continue
		}
		go handleConnection(inputConn, outputPath, *ratio, *clean)
	}
}
