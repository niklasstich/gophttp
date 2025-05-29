package http

import (
	"bufio"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
)

type Request struct {
	Method
	Version
	Path    string
	Headers map[string]Header
	Body    []byte
}

type Header struct {
	Name  string
	Value string
}

var ErrInvalidRequest = fmt.Errorf("invalid HTTP request format")

func ParseRequest(conn net.Conn) (Request, error) {
	//wrap connection in a buffered scanner
	r := bufio.NewReader(conn)
	s := bufio.NewScanner(r)

	request := Request{}
	buf := []string{}
	for s.Scan() {
		line := s.Text()
		if line == "" {
			break
		}
		buf = append(buf, line+"\n")
	}
	//parse method, path, version and headers
	if len(buf) == 0 {
		//invalid request
		return Request{}, ErrInvalidRequest
	}
	firstLineParts := strings.Split(buf[0], " ")
	if len(firstLineParts) < 3 {
		return Request{}, ErrInvalidRequest
	}
	var err error
	request.Method, err = parseMethod(firstLineParts[0])
	if err != nil {
		return Request{}, fmt.Errorf("unable to parse request: %w", err)
	}
	request.Path = firstLineParts[1]
	request.Version, err = parseVersion(firstLineParts[2])
	if err != nil {
		return Request{}, fmt.Errorf("unable to parse request: %w", err)
	}

	//parse headers
	request.Headers = make(map[string]Header)
	//we assume every line after the first line is a header
	//until we find an empty line
	for _, line := range buf[1:] {
		if line == "\n" {
			break
		}
		s := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(s[0])
		value := strings.TrimSpace(s[1])
		request.Headers[name] = Header{
			name, value,
		}
	}

	//determine if we need to look for a body too
	//refuse reading body on GET, HEAD, OPTIONS, CONNECT and TRACE
	if request.Method == GET ||
		request.Method == HEAD ||
		request.Method == OPTIONS ||
		request.Method == CONNECT ||
		request.Method == TRACE {
		return request, nil
	}

	for _, header := range request.Headers {
		//if we have a Content-Length header, we read the expected length of bytes as the body
		if header.Name == "Content-Length" {
			handleContentLength(&request, r)
			break
		} else if header.Name == "Transfer-Encoding" {
			handleTransferEncoding(&request, s)
			break
		}
	}

	return request, nil

}

func handleContentLength(request *Request, r *bufio.Reader) error {
	bodyLen, err := strconv.Atoi(request.Headers["Content-Length"].Value)
	if err != nil {
		return fmt.Errorf("could not parse Content-Length: %v", err)
	}

	//force read bodyLen bytes from the connection
	buffer := make([]byte, bodyLen)
	n, err := r.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}
	if n != bodyLen {
		return fmt.Errorf("unexpected body length %d instead of %d", n, bodyLen)
	}

	decoded, err := handleContentEncoding(buffer, request)
	if err != nil {
		return fmt.Errorf("failed to decode body: %v", err)
	}

	request.Body = decoded
	return nil
}

func handleTransferEncoding(request *Request, s *bufio.Scanner) error {
	buffer := make([]byte, 0)
	//loop until we read a zero and an empty line (end of body)
	for s.Scan() {
		//read length of next block
		blockLen := s.Text()

		bodyLen, err := strconv.ParseInt(blockLen, 16, 64)
		if err != nil {
			return fmt.Errorf("failed to read transfer encoding block length: %v", err)
		}

		if bodyLen == 0 {
			b, err := readChunk(s, bodyLen)
			if err != nil {
				return err
			}
			if len(b) == 0 {
				//done reading
				break
			} else {
				return fmt.Errorf("received 0 block length but following line was not empty")
			}
		}

		b, err := readChunk(s, bodyLen)
		if err != nil {
			return err
		}

		//append b to buffer
		buffer = slices.Concat(buffer, b)
	}

	//decode
	buffer, err := handleContentEncoding(buffer, request)
	if err != nil {
		return err
	}

	request.Body = buffer

	return nil
}

func readChunk(s *bufio.Scanner, bodyLen int64) ([]byte, error) {
	if !s.Scan() {
		return nil, fmt.Errorf("unexpected end of body")
	}
	err := s.Err()
	if err != nil {
		return nil, err
	}
	retval := s.Bytes()
	if int64(len(retval)) != bodyLen {
		return nil, fmt.Errorf("unexpected body length %d instead of %d", len(retval), bodyLen)
	}

	return retval, nil
}

func handleContentEncoding(data []byte, request *Request) ([]byte, error) {
	//TODO: support content encoding in requests?
	return data, nil
}

func parseMethod(method string) (Method, error) {
	switch method {
	case "GET":
		return GET, nil
	case "HEAD":
		return HEAD, nil
	case "POST":
		return POST, nil
	case "PUT":
		return PUT, nil
	case "DELETE":
		return DELETE, nil
	case "CONNECT":
		return CONNECT, nil
	case "OPTIONS":
		return OPTIONS, nil
	case "TRACE":
		return TRACE, nil
	case "PATCH":
		return PATCH, nil
	default:
		return 0, fmt.Errorf("unknown HTTP method: %s", method)
	}
}

func parseVersion(version string) (Version, error) {
	switch strings.TrimSpace(version) {
	case "HTTP/1.0":
		return HTTP1_0, nil
	case "HTTP/1.1":
		return HTTP1_1, nil
	case "HTTP/2.0":
		return HTTP2, nil
	case "HTTP/3.0":
		return HTTP3, nil
	default:
		return 0, fmt.Errorf("unknown HTTP version: %s", version)
	}
}
