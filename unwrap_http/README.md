# Unwrap HTTP

Is a tool that listens for connections on its `$INPUT` (stream Unix socket) and proxies them to its `$OUTPUT` (stream Unix socket) in such a way that:
* `$INPUT` -> `$OUTPUT` direction gets its HTTP headers stripped, leaving only the body (without any limits by Content-Length and such, just the raw TCP stream)
* `$OUTPUT` -> `$INPUT` direction gets HTTP response headers prepended before the data gets proxied
