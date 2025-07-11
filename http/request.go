package http

import (
	"bufio"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Method
	Version
	Path    string
	Headers Headers
	Body    []byte
}

var ErrInvalidRequest = fmt.Errorf("invalid HTTP request format")
var ErrInvalidHttpMethod = fmt.Errorf("invalid HTTP method")
var ErrInvalidHttpVersion = fmt.Errorf("invalid HTTP version")

type errInvalidHttpMethod struct {
	Method string
}

func (e errInvalidHttpMethod) Error() string {
	return fmt.Sprintf("invalid HTTP method: %s", e.Method)
}

func (e errInvalidHttpMethod) Unwrap() error {
	return ErrInvalidHttpMethod
}

type errInvalidHttpVersion struct {
	Version string
}

func (e errInvalidHttpVersion) Unwrap() error {
	return ErrInvalidHttpVersion
}

func (e errInvalidHttpVersion) Error() string {
	return fmt.Sprintf("invalid HTTP version: %s", e.Version)
}

func ParseRequest(ctx Context, r *bufio.Reader, s *bufio.Scanner) (*Request, error) {
	//set a 5s read timeout on the underlying connection
	err := ctx.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("couldn't set read deadline on conn when parsing request")
	}

	request := &Request{}
	buf, err := readRequestLineAndHeaders(ctx, s)
	if err != nil {
		return nil, err
	}

	// parse method, path, version and headers
	err = parseRequestLineAndHeaders(buf, request, ctx)
	if err != nil {
		return nil, err
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
			err = handleContentLength(request, r)
			break
		} else if header.Name == "Transfer-Encoding" {
			err = handleTransferEncoding(request, s)
			break
		}
	}
	if err != nil {
		ctx.AdditionalData["BadRequestReason"] = "Failed parsing request body"
		return nil, fmt.Errorf("failed parsing request body: %w", err)
	}

	return request, nil

}

func readRequestLineAndHeaders(ctx Context, s *bufio.Scanner) ([]string, error) {
	buf := []string{}
	//we already advanced the scanner by one token when checking if there was another request in the conn
	//so we must read the first token right away before calling s.Scan again
	for {
		line := s.Text()
		if line == "" {
			break
		}
		buf = append(buf, line+"\n")
		if !s.Scan() {
			break
		}
	}
	if len(buf) == 0 {
		//invalid request
		ctx.AdditionalData["BadRequestReason"] = "Empty request"
		return nil, fmt.Errorf("%w: empty buffer", ErrInvalidRequest)
	}
	return buf, nil
}

func parseRequestLineAndHeaders(buf []string, request *Request, ctx Context) error {
	//parse method, path, version and headers
	firstLineParts := strings.Split(buf[0], " ")
	if len(firstLineParts) < 3 {
		ctx.AdditionalData["BadRequestReason"] = "Invalid request format"
		return fmt.Errorf("%w: request line is %d parts long", ErrInvalidRequest, len(firstLineParts))
	}
	var err error
	request.Method, err = parseMethod(firstLineParts[0])
	if err != nil {
		ctx.AdditionalData["BadRequestReason"] = "Invalid HTTP method"
		return fmt.Errorf("unable to parse request: %w", err)
	}
	request.Path = firstLineParts[1]
	request.Version, err = parseVersion(firstLineParts[2])
	if err != nil {
		ctx.AdditionalData["BadRequestReason"] = "Invalid HTTP version"
		return fmt.Errorf("unable to parse request: %w", err)
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
	return nil
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
		return 0, errInvalidHttpMethod{method}
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
		return 0, errInvalidHttpVersion{version}
	}
}
