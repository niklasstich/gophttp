package http

type Method int
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
