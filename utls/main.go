package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	utls "github.com/refraction-networking/utls"
)

var fingerprints = map[string]utls.ClientHelloID{
	"Golang":           utls.HelloGolang,
	"Randomized":       utls.HelloRandomized,
	"RandomizedALPN":   utls.HelloRandomizedALPN,
	"RandomizedNoALPN": utls.HelloRandomizedNoALPN,

	"Firefox_Auto": utls.HelloFirefox_Auto,
	"Firefox_55":   utls.HelloFirefox_55,
	"Firefox_56":   utls.HelloFirefox_56,
	"Firefox_63":   utls.HelloFirefox_63,
	"Firefox_65":   utls.HelloFirefox_65,
	"Firefox_99":   utls.HelloFirefox_99,
	"Firefox_102":  utls.HelloFirefox_102,
	"Firefox_105":  utls.HelloFirefox_105,
	"Firefox_120":  utls.HelloFirefox_120,

	"Chrome_Auto":                 utls.HelloChrome_Auto,
	"Chrome_58":                   utls.HelloChrome_58,
	"Chrome_62":                   utls.HelloChrome_62,
	"Chrome_70":                   utls.HelloChrome_70,
	"Chrome_72":                   utls.HelloChrome_72,
	"Chrome_83":                   utls.HelloChrome_83,
	"Chrome_87":                   utls.HelloChrome_87,
	"Chrome_96":                   utls.HelloChrome_96,
	"Chrome_100":                  utls.HelloChrome_100,
	"Chrome_102":                  utls.HelloChrome_102,
	"Chrome_106_Shuffle":          utls.HelloChrome_106_Shuffle,
	"Chrome_100_PSK":              utls.HelloChrome_100_PSK,
	"Chrome_112_PSK_Shuf":         utls.HelloChrome_112_PSK_Shuf,
	"Chrome_114_Padding_PSK_Shuf": utls.HelloChrome_114_Padding_PSK_Shuf,
	"Chrome_115_PQ":               utls.HelloChrome_115_PQ,
	"Chrome_115_PQ_PSK":           utls.HelloChrome_115_PQ_PSK,
	"Chrome_120":                  utls.HelloChrome_120,
	"Chrome_120_PQ":               utls.HelloChrome_120_PQ,
	"Chrome_131":                  utls.HelloChrome_131,
	"Chrome_133":                  utls.HelloChrome_133,

	"IOS_Auto": utls.HelloIOS_Auto,
	"IOS_11_1": utls.HelloIOS_11_1,
	"IOS_12_1": utls.HelloIOS_12_1,
	"IOS_13":   utls.HelloIOS_13,
	"IOS_14":   utls.HelloIOS_14,

	"Android_11_OkHttp": utls.HelloAndroid_11_OkHttp,

	"Edge_Auto": utls.HelloEdge_Auto,
	"Edge_85":   utls.HelloEdge_85,
	"Edge_106":  utls.HelloEdge_106,

	"Safari_Auto": utls.HelloSafari_Auto,
	"Safari_16_0": utls.HelloSafari_16_0,

	"360_Auto": utls.Hello360_Auto,
	"360_7_5":  utls.Hello360_7_5,
	"360_11_0": utls.Hello360_11_0,

	"QQ_Auto": utls.HelloQQ_Auto,
	"QQ_11_1": utls.HelloQQ_11_1,
}

func main() {
	sni := flag.String("sni", "", "server name to send in the TLS handshake")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("usage: utls [-sni server_name] <fingerprint>")
	}
	fingerprint, ok := fingerprints[flag.Arg(0)]
	if !ok {
		log.Fatalf("unknown fingerprint %q", flag.Arg(0))
	}
	inputPath, outputPath := os.Getenv("INPUT"), os.Getenv("OUTPUT")
	if inputPath == "" || outputPath == "" {
		log.Fatalf("INPUT and OUTPUT environment variables must be set")
	}

	listener, err := listenUnix(inputPath)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer os.Remove(inputPath)
	defer listener.Close()

	var connections sync.WaitGroup
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		plain, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				break
			}
			log.Fatalf("accept: %v", err)
		}
		connections.Add(1)
		go func() {
			defer connections.Done()
			proxy(plain, outputPath, *sni, fingerprint)
		}()
	}

	stop()
	connections.Wait()
}

func listenUnix(path string) (net.Listener, error) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return net.Listen("unix", path)
}

func proxy(plain net.Conn, outputPath, sni string, fingerprint utls.ClientHelloID) {
	defer plain.Close()

	raw, err := net.Dial("unix", outputPath)
	if err != nil {
		log.Printf("dial %s: %v", outputPath, err)
		return
	}
	defer raw.Close()

	tlsConn := utls.UClient(raw, &utls.Config{ServerName: sni, InsecureSkipVerify: true}, fingerprint)
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("handshake: %v", err)
		return
	}

	var copyDone sync.WaitGroup
	copyDone.Add(1)
	go func() {
		defer copyDone.Done()
		io.Copy(tlsConn, plain)
		closeWrite(tlsConn)
	}()
	io.Copy(plain, tlsConn)
	closeWrite(plain)
	copyDone.Wait()
}

func closeWrite(conn net.Conn) {
	if writer, ok := conn.(interface{ CloseWrite() error }); ok {
		writer.CloseWrite()
	}
}
