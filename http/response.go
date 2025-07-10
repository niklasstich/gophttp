package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

type Response struct {
	Status
	Headers Headers
	Body    interface{}
}

func NewResponse() *Response {
	return &Response{Headers: make(map[string]Header)}
}

// AddHeader adds the given header to the response, overwriting any header that might be present already for the given key
func (r Response) AddHeader(header Header) {
	r.Headers[header.Name] = header
}

var ErrUnknownBodyType = fmt.Errorf("unknown body type")

func (r Response) WriteToConn(conn net.Conn) error {
	w := bufio.NewWriter(conn)
	_, err := w.WriteString(fmt.Sprintf("HTTP/1.1 %s\n", r.Status))
	if err != nil {
		return err
	}
	//write all headers
	for _, header := range r.Headers {
		_, err = w.WriteString(fmt.Sprintf("%s: %s\n", header.Name, strings.TrimRight(header.Value, "\n")))
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
	} else if b, ok := r.Body.(bytes.Buffer); ok {
		_, err := w.Write(b.Bytes())
		if err != nil {
			return err
		}
	} else if c, ok := r.Body.(<-chan StreamedResponseChunk); ok {
		//if we get a byte slice channel, start a loop where we read from said channel until it closes
		//we block here and do not create another goroutine because we need to wait until we fully wrote our response
		//before moving on to the next request in the TCP connection
		for {
			select {
			case chunk, more := <-c:
				if !more {
					return w.Flush()
				}
				if chunk.Err != nil {
					return chunk.Err
				}
				_, err := w.Write(chunk.Data)
				if err != nil {
					return err
				}
			case <-time.After(15 * time.Second): //TODO: make configurable
				return fmt.Errorf("read timeout on body channel")
			}
		}
	} else {
		//log and return err (500)
		return fmt.Errorf("%v: %T", ErrUnknownBodyType, r.Body)
	}

	return w.Flush()
}
