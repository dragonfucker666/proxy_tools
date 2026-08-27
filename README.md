# Proxy Tools

are small proxy programs to be combined together into complete proxy programs that fit your needs. You can create your own tools and combine them with tools of others

~~This repository could have been released earlier, but instead of making it, I was spending time swimming under a "no swimming" sign~~

## Conventions

### Programming

* If your program accepts new connections from other tools, make it accept them on a Unix socket with the path taken from "INPUT" environment variable
* If your program makes its own connections to other tools, make it do them into a Unix socket with the path taken from "OUTPUT" environment variable

Break big ideas into many small tools if possible

Accept environment variables for cooperation with other programs, and command line arguments for independent parameters

### Running

Every combination is ran from a directory filled with tools, with a subdirectory for socket storage for each parallel run (so they won't collide)

## Usage

### Once

* Execute `./build_all.sh` (you'd need to install `go` (Golang's source code manager))
* Add your own tools to `build/`

### Every time

* Go into `build/`
* Run your tools
* If you use the "chain" tool, don't forget to specify a `--storage` directory for all the sockets of the run

## Example configurations

### "Graceful Xi"

Simple socket-to-socket TCP transport based mostly on well-established web technologies. Avoids TLS-in-TLS detection, simultaneous connection limits, unusual fingerprint detection and active probing. Server connections are encrypted

Server (under a TLS terminator and http/2 demultiplexer reverse proxy such as NGINX):

```
./chain --storage graceful_xi + ./listen_tcp 127.0.0.1 "$PORT_THE_REVERSE_PROXY_OUTPUTS_TO" + ./unwrap_http + ./cleaner + ./send_tcp 127.0.0.1 "$DESTINATION_PORT"
```

or

```
./chain --storage graceful_xi + env INPUT=./reverse_proxy_socket_path ./unwrap_http + env OUTPUT=./next_program_socket_path ./cleaner
```

Client (raw, listens on a socket):

```
./chain --storage graceful_xi + ./listen_tcp 127.0.0.1 "$LOCAL_PORT_FOR_INCOMING_CONNECTIONS" + ./trasher + ./mux_http2 --scheme https --authority example.com --path /secret-path-choose-yourself + ./utls --sni example.com --fingerprint Firefox_Auto + ./send_tcp example.com 443
```
