<p align="center">
  <img src="media/logo.png" alt="gophttp logo" width="200"/>
</p>

# gophttp

A fast, minimal, and extensible HTTP/1.1 server written in Go. Written as a hobby project to learn more about HTTP and Go.

---

## Features (Implemented)

- **Static File Serving:** Serves static files and directories with template support.
- **Handler Collection per Path:** Register handlers for different HTTP methods on each route.
- **Common Response Headers:** Automatic writing of common headers on every response.
- **Compression:** Brotli support for static content (see TODO for details).
- **Connection Keep-Alive:** Supports `Connection: keep-alive` for persistent connections.
- **Chunked Transfer Encoding:** Supports chunked transfer responses using Go channels.
- **Radix Tree Routing:** Efficient path matching using a custom radix tree implementation.

## Warning
⚠️ This server is a hobby project and as such is NOT fully HTTP/1.1 compliant (yet)! It also is NOT hardened against 
even the common attacks against HTTP servers and as such ***SHOULD NOT BE USED IN PRODUCTION OR ANYTHING FACING THE PUBLIC INTERNET.*** You have been warned.⚠️

## Getting Started

1. **Clone the repo:**
```sh
git clone https://github.com/yourusername/gophttp.git
cd gophttp
```
2. **Build and run:**
```sh
go run main.go
```
3. **Test:**
```sh
go test -tags test ./...
```

You can either change main.go to fit your requirements or write your own implementation using `HttpServer`.

## Roadmap / TODO

See [`TODO.md`](TODO.md) for a detailed list of planned features and improvements.

## License

See [`LICENCE.md`](LICENCE.md).

