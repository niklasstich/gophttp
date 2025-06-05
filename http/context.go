package http

import (
	"net"
)

type Context struct {
	*Request
	*Response
	net.Conn
	AdditionalData map[string]interface{}
}

func NewContext(conn net.Conn) Context {
	return Context{AdditionalData: make(map[string]interface{}), Conn: conn}
}
