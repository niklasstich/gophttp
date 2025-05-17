package http

//go:generate stringer -type=Method
type Method int

//go:generate stringer -type=Version
type Version int

const (
	GET Method = iota
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
)

const (
	HTTP1_0 Version = iota
	HTTP1_1
	HTTP2
	HTTP3
)
