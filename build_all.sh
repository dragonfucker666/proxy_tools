build_go() {
    (
	cd "$1"
	go build main.go
    )
    mv "$1/main" "build/$(basename "$1")"
}

mkdir -p build/
build_go chain/
build_go listen_tcp/
build_go mux_http2/
build_go send_tcp/
build_go trasher/
build_go unwrap_http/
build_go utls/
