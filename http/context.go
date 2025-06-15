package http

import (
	"net"
)

type Context struct {
	*Request
	*Response
	net.Conn
	Index          uint64
	AdditionalData map[string]interface{}
}

func NewContext(conn net.Conn, index uint64) Context {
	return Context{AdditionalData: make(map[string]interface{}), Conn: conn, Index: index, Response: NewResponse()}
}
