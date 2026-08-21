# Mux HTTP/2

Accepts these flags:
* `--scheme` (http/2 pseudo-header; default="https")
* `--authority` (http/2 pseudo-header; default="example.com")
* `--path` (http/2 pseudo-header; default="/")

It creates one connection to the `$OUTPUT` Unix stream socket, accepts connections from `$INPUT` Unix stream socket, and creates POST HTTP/2 requests with the aforementioned pseudo-headers along that one connection to the `$OUTPUT`

If HTTP/2 breaks (for example, if the `$OUTPUT` connection closes, or if it starts sending errorneous data back) when establishing a new connection inside the channel, one reconnection to `$OUTPUT` is made, but no more than one (to enable reconnects, but not spam)
