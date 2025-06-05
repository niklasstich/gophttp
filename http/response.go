package http

import (
	"bufio"
	"fmt"
	"net"
)

type Response struct {
	Status
	Headers Headers
	Body    interface{}
}
type Headers map[string]Header

func (m Headers) HasHeader(key string) bool {
	for _, header := range m {
		if header.Name == key {
			return true
		}
	}
	return false
}

func NewResponse() *Response {
	return &Response{Headers: make(map[string]Header)}
}

// AddHeader adds the given header to the response, overwriting any header that might be present already for the given key
func (r Response) AddHeader(header Header) {
	r.Headers[header.Name] = header
}

func (r Response) WriteToConn(conn net.Conn) error {
	w := bufio.NewWriter(conn)
	_, err := w.WriteString(fmt.Sprintf("HTTP/1.1 %s\n", r.Status))
	if err != nil {
		return err
	}
	//write all headers
	for _, header := range r.Headers {
		_, err = w.WriteString(fmt.Sprintf("%s: %s\n", header.Name, header.Value))
		if err != nil {
			return err
		}
	}
	_, err = w.WriteString("\n")
	if err != nil {
		return err
	}

	//if resp has no body, we are done
	if r.Body == nil {
		return w.Flush()
	}

	//write body depending on the type of data
	if s, ok := r.Body.(string); ok {
		_, err = w.WriteString(s)
		if err != nil {
			return err
		}
	} else if b, ok := r.Body.([]byte); ok {
		_, err = w.Write(b)
		if err != nil {
			return err
		}
	} else {
		//don't know how to write it, panic for now
		//TODO: fix this
		panic("unknown body type")
	}

	return w.Flush()
}
