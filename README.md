# Proxy Tools

is an architecture (and an example set of programs) intended to help with building proxying programs collectively

## Proxy Tools protocol

Every Proxy Tool is just a binary that can accept any of these flags on the command line, which are guaranteed to go before positional arguments:

* `--output ./output_socket_path` - a path to the output Unix file socket
* `--input ./input_socket_path` - a path to the input Unix file socket

To be combinable, programs are expected to listen for connections on the input socket and make connections to the output socket. Usually 1:1, although some programs may have more input connections than output connections, or more output connections than input connections (for example, when multiplexing or demultiplexing connections)

You don't have to check for the output socket's existence, or whether the input socket file exists; you may if you want to, but that's not important. Wrapper programs usually take care of deleting the input socket and making sure the output socket exists before running the program

Programs should create their socket input files at startup if an input is required

It is not necessary but recommended to delete the input socket on exit (to not create clutter)

## Problem Proxy Tools solves

Proxy combines like [xray-core](https://github.com/XTLS/Xray-core/) are monolithic, and their configurations are very interconnected, making them hard to understand and extend

Proxy Tools are built around the idea that every stage in the proxying process can be handled by a separate program

## Disadvantages over proxy combines

* Slightly lower performance

## Advantages over proxy combines

* Easy to extend with own modules
* Easy to understand configuration
* Uses more OS-native technologies, so requires less prerequisite knowledge
* Easier to mix and match different modules and test them in isolation

## How to use this repo

Every subdirectory in here is its own tool to be combined, except for the "chain" tool: it is intended to tie together other tools, it provides simple socket file management as well

"chain" receives the subprogram separator as its first argument, and then the programs to tie together:

```
chain + program_that_listens arg1 arg2 + ./middle_program arg1 arg2 --key value + program_that_sends --flag
```

"chain" prepends `--input` and `--output` arguments when calling subprograms, and only calls subprograms when their output is ready to use and the input was deleted

### Example configurations

#### "Graceful Xi"

Simple socket-to-socket TCP transport based mostly on well-established web technologies. Avoids TLS-in-TLS detection, simultaneous connection limits, unusual fingerprint detection and active probing. Server connections are encrypted

Server (under a TLS terminator and http/2 demultiplexer reverse proxy such as NGINX):

```
chain + listen_tcp --host 127.0.0.1 --port "$PORT_THE_REVERSE_PROXY_OUTPUTS_TO" + cleaner + send_tcp --host 127.0.0.1 --port "$DESTINATION_PORT"
```

Client (raw, listens on a socket):

```
chain + listen_tcp --host 127.0.0.1 --port "$LOCAL_PORT_FOR_INCOMING_CONNECTIONS" + trasher + http2mux --host example.com --path /secret-path-choose-yourself + utls --fingerprint firefox-latest + send_tcp --host example.com --port 443
```
