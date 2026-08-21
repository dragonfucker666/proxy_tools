// Command mux-http2 accepts connections from the $INPUT Unix socket and
// forwards each one as a POST HTTP/2 request over a single shared connection
// to the $OUTPUT Unix socket.
//
// The request body is read from the incoming connection and the response body
// is written back to it. If HTTP/2 breaks, the shared connection to $OUTPUT is
// re-established once per request before giving up.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/http2"
)

func main() {
	scheme := flag.String("scheme", "https", "HTTP/2 :scheme pseudo-header")
	authority := flag.String("authority", "example.com", "HTTP/2 :authority pseudo-header")
	path := flag.String("path", "/", "HTTP/2 :path pseudo-header")
	flag.Parse()

	input := os.Getenv("INPUT")
	output := os.Getenv("OUTPUT")
	if input == "" || output == "" {
		log.Fatal("INPUT and OUTPUT environment variables must be set")
	}

	listener := listen(input)
	defer listener.Close()
	defer os.Remove(input)

	// The transport keeps a single HTTP/2 connection to $OUTPUT and shares it
	// between all requests.
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, _, _ string, _ *tls.Config) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", output)
		},
	}

	log.Printf("listening on %s, forwarding to %s", input, output)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go forward(transport, *scheme, *authority, *path, conn)
	}
}

func listen(path string) net.Listener {
	os.Remove(path) // drop a stale socket left over from a previous run
	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}
	return listener
}

// forward turns one $INPUT connection into one POST request: the connection is
// used as the request body and the response body is written back into it.
func forward(transport *http2.Transport, scheme, authority, path string, conn net.Conn) {
	defer conn.Close()

	endpoint := url.URL{Scheme: scheme, Host: authority, Path: path}
	// Strip the Close method: the transport closes the request body when the
	// request ends, which must not close the $INPUT connection itself.
	req, err := http.NewRequest(http.MethodPost, endpoint.String(), struct{ io.Reader }{conn})
	if err != nil {
		log.Print(err)
		return
	}

	var resp *http.Response
	for attempt := 1; attempt <= 2; attempt++ {
		resp, err = transport.RoundTrip(req)
		if err == nil {
			break
		}
		// HTTP/2 broke: force a fresh connection for the retry, but only one.
		transport.CloseIdleConnections()
	}
	if err != nil {
		log.Printf("POST %s failed: %v", endpoint.String(), err)
		return
	}
	defer resp.Body.Close()

	io.Copy(conn, resp.Body)
}
