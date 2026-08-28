# Listen TCP

is a tool for listening on some host (first argument; may be a domain name or an ipv4/ipv6 address) and port (second argument) and proxying all the connections into the `$OUTPUT` Unix file socket (`$OUTPUT` contains the path to the socket)

The `--ip` flag specifies the IP network to use: `v4`, `v6` or "any". Default is "any"
