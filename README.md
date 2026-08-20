# Proxy Tools

is a bunch of programs and a shared protocol to create modular proxies, like [xray-core](https://github.com/XTLS/Xray-core/) but more customizable and decoupled, much easier to understand and use. You can write proxy tools yourself and mix them with tools made by other people to create pipelines you need

Every subdirectory in here is a proxy tool. Build them if you need them. You just need the executables

## Protocol

Every "tool" is just an executable file. If you want to be able to tie your tools with other tools, you can make your tools accept these flagged arguments (guaranteed to come before other arguments):

* `--input ./input_socket_path` - Unix file socket the program needs to create and listen for connections from
* `--output ./output_socket_path` - Unix file socket the program needs to make connections into and write output data into those connections

Tools can be tied together using the "chain" tool

## Example configurations

### "Graceful Xi"

Simple socket-to-socket TCP transport based mostly on well-established web technologies. Avoids TLS-in-TLS detection, simultaneous connection limits, unusual fingerprint detection and active probing. Server connections are encrypted

Server (under a TLS terminator and http/2 demultiplexer reverse proxy such as NGINX):

```
chain + listen_tcp --host 127.0.0.1 --port "$PORT_THE_REVERSE_PROXY_OUTPUTS_TO" + cleaner + send_tcp --host 127.0.0.1 --port "$DESTINATION_PORT"
```

Client (raw, listens on a socket):

```
chain + listen_tcp --host 127.0.0.1 --port "$LOCAL_PORT_FOR_INCOMING_CONNECTIONS" + trasher + http2mux --host example.com --path /secret-path-choose-yourself + utls --fingerprint firefox-latest + send_tcp --host example.com --port 443
```
