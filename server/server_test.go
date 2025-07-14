//go:build test

package server_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"gophttp/handlers"
	"gophttp/http"
	"gophttp/server"
)

func TestCustomAddedHandlerIsCalled(t *testing.T) {
	// Start server on a random port
	port := 8089 // Use a test port unlikely to be in use
	httpServer := server.NewHttpServer(port)

	// Register a custom handler
	const testPath = "/test"
	const expectedBody = "Hello, test!"
	err := httpServer.AddHandler(testPath, http.GET, handlers.HandlerFunc(func(ctx http.Context) error {
		ctx.Response.Status = http.StatusOK
		ctx.Response.Body = expectedBody
		ctx.Response.AddHeader(http.Header{Name: "Content-Type", Value: "text/plain"})
		return nil
	}))
	if err != nil {
		t.Fatalf("failed setting up handler: %v", err)
	}
	ctx, cfunc := context.WithCancel(context.Background())
	servClosed := make(chan bool, 1)

	// Start server in background
	go func() {
		defer func() { servClosed <- true }()
		err := httpServer.StartServing(ctx)
		if err != nil {
			t.Errorf("failed to serve: %v", err)
			return
		}
	}()
	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Connect to server
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Send HTTP GET request
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: localhost\r\n\r\n", testPath)
	_, err = conn.Write([]byte(req))
	if err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	var response strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		response.WriteString(line)
		if line == "\r\n" || line == "\n" {
			break // End of headers
		}
	}
	// Read body
	body, _ := reader.ReadString('\n')
	response.WriteString(body)

	if !strings.Contains(response.String(), expectedBody) {
		t.Errorf("expected body %q in response, got: %q", expectedBody, response.String())
	}
	cfunc()
	//block until server is dead (this is to prevent having multiple tests that create servers running at the same time)
	//((assuming that all tests run sequentially not parallel, which seems to be the case))
	_ = <-servClosed
}

func TestStreamedResponseWithDelay(t *testing.T) {
	port := 8090 // Use a different test port
	httpServer := server.NewHttpServer(port)

	const testPath = "/stream"
	seg1 := []byte("segment1-")
	seg2 := []byte("segment2!")

	err := httpServer.AddHandler(testPath, http.GET, handlers.HandlerFunc(func(ctx http.Context) error {
		ctx.Response.Status = http.StatusOK
		ctx.Response.AddHeader(http.Header{Name: "Content-Type", Value: "text/plain"})
		ch := make(chan http.StreamedResponseChunk)
		ctx.Response.Body = ch
		go func() {
			ch <- http.StreamedResponseChunk{Data: seg1}
			time.Sleep(1 * time.Second)
			ch <- http.StreamedResponseChunk{Data: seg2}
			close(ch)
		}()
		return nil
	}))
	if err != nil {
		t.Fatalf("failed setting up handler: %v", err)
	}
	ctx, cfunc := context.WithCancel(context.Background())
	servClosed := make(chan bool, 1)

	go func() {
		defer func() { servClosed <- true }()
		err := httpServer.StartServing(ctx)
		if err != nil {
			t.Errorf("failed to serve: %v", err)
			return
		}
	}()
	time.Sleep(200 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: localhost\r\n\r\n", testPath)
	_, err = conn.Write([]byte(req))
	if err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	reader := bufio.NewReader(conn)
	var response strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		response.WriteString(line)
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	// Read first segment
	body1 := make([]byte, len(seg1)+4)
	_, err = reader.Read(body1)
	if err != nil {
		t.Fatalf("failed to read first segment: %v", err)
	}
	// Wait for the second segment (should be delayed)
	body2 := make([]byte, len(seg2)+4)
	_, err = reader.Read(body2)
	if err != nil {
		t.Fatalf("failed to read second segment: %v", err)
	}
	response.Write(body1)
	response.Write(body2)

	body := response.String()

	if !strings.Contains(body, "9\r\n"+string(seg1)) || !strings.Contains(body, "9\r\n"+string(seg2)) {
		t.Errorf("expected streamed segments in response, got: %q", body)
	}

	if !strings.Contains(body, "Transfer-Encoding: chunked") {
		t.Error("expected 'Transfer-Encoding: chunked' header but didn't find it")
	}
	cfunc()
	_ = <-servClosed
}

func TestStreamedResponseWithDelayAndBrotliAccept(t *testing.T) {
	timeFunc := func() time.Time {
		return time.Date(2025, 7, 13, 11, 57, 50, 0, time.UTC)
	}

	handlers.SetTimeFunc(timeFunc)

	port := 8091 // Use a different test port
	httpServer := server.NewHttpServer(port)

	const testPath = "/stream-brotli"
	seg1 := bytes.Repeat([]byte("L"), 256)
	seg2 := []byte("\n\n")
	seg3 := bytes.Repeat([]byte("F"), 256)

	handlerFn := handlers.HandlerFunc(func(ctx http.Context) error {
		ctx.Response.Status = http.StatusOK
		ctx.Response.AddHeader(http.Header{Name: "Content-Type", Value: "text/plain"})
		ch := make(chan http.StreamedResponseChunk)
		ctx.Response.Body = ch
		go func() {
			defer close(ch)
			ch <- http.StreamedResponseChunk{Data: seg1}
			time.Sleep(100 * time.Millisecond)
			ch <- http.StreamedResponseChunk{Data: seg2}
			time.Sleep(100 * time.Millisecond)
			ch <- http.StreamedResponseChunk{Data: seg3}
		}()
		return nil
	})
	compressionHandler := handlers.NewCompressionHandler()
	handler := handlers.ComposeHandlers(handlerFn, compressionHandler)

	err := httpServer.AddHandler(testPath, http.GET, handler)
	if err != nil {
		t.Fatalf("failed setting up handler: %v", err)
	}
	ctx, cfunc := context.WithCancel(context.Background())
	servClosed := make(chan bool, 1)

	go func() {
		defer func() { servClosed <- true }()
		err := httpServer.StartServing(ctx)
		if err != nil {
			t.Errorf("failed to serve: %v", err)
			return
		}
	}()
	time.Sleep(200 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: localhost\r\nAccept-Encoding: br\r\n\r\n", testPath)
	_, err = conn.Write([]byte(req))
	if err != nil {
		t.Fatalf("failed to write request: %v", err)
	}

	reader := bufio.NewReader(conn)
	//brotli buffers all segments before compressing, meaning we only get one big compressed chunk
	body1 := make([]byte, 203)
	_, err = reader.Read(body1)
	if err != nil {
		t.Fatalf("failed to read first segment: %v", err)
	}
	expectedBytes := []byte{
		0x48, 0x54, 0x54, 0x50, 0x2f, 0x31, 0x2e, 0x31, 0x20, 0x32, 0x30, 0x30, 0x20, 0x4f, 0x4b, 0x0a,
		0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x20, 0x6b, 0x65, 0x65, 0x70,
		0x2d, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x0a, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x2d, 0x45,
		0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x3a, 0x20, 0x62, 0x72, 0x0a, 0x43, 0x6f, 0x6e, 0x74,
		0x65, 0x6e, 0x74, 0x2d, 0x54, 0x79, 0x70, 0x65, 0x3a, 0x20, 0x74, 0x65, 0x78, 0x74, 0x2f, 0x70,
		0x6c, 0x61, 0x69, 0x6e, 0x0a, 0x44, 0x61, 0x74, 0x65, 0x3a, 0x20, 0x53, 0x75, 0x6e, 0x2c, 0x20,
		0x31, 0x33, 0x20, 0x4a, 0x75, 0x6c, 0x20, 0x32, 0x30, 0x32, 0x35, 0x20, 0x31, 0x31, 0x3a, 0x35,
		0x37, 0x3a, 0x35, 0x30, 0x20, 0x47, 0x4d, 0x54, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x3a,
		0x20, 0x67, 0x6f, 0x70, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x30, 0x2e, 0x31, 0x0a, 0x54, 0x72, 0x61,
		0x6e, 0x73, 0x66, 0x65, 0x72, 0x2d, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x3a, 0x20,
		0x63, 0x68, 0x75, 0x6e, 0x6b, 0x65, 0x64, 0x0a, 0x0a, 0x31, 0x31, 0x0d, 0x0a, 0x1b, 0x01, 0x02,
		0x00, 0x24, 0x15, 0x8c, 0x98, 0x6a, 0xb1, 0xcd, 0x0a, 0x40, 0xe4, 0x3e, 0x47, 0x00, 0x0d, 0x0a,
		0x30, 0x0d, 0x0a, 0x0d, 0x0a,
	}

	if len(body1) < len(expectedBytes) {
		t.Errorf("expected len %d but body is only %d long", len(expectedBytes), len(body1))
	}
	for i, ex := range expectedBytes {
		if ex != body1[i] {
			t.Errorf("expected byte %d but got %d at idx %d", ex, body1[i], i)
		}
	}

	cfunc()
	_ = <-servClosed
}

func TestKeepAliveSupport(t *testing.T) {
	port := 8092 // Use a different test port
	httpServer := server.NewHttpServer(port)

	const testPath = "/test"
	const expectedBody = "Hello, test!"
	err := httpServer.AddHandler(testPath, http.GET, handlers.HandlerFunc(func(ctx http.Context) error {
		ctx.Response.Status = http.StatusOK
		ctx.Response.Body = expectedBody
		ctx.Response.AddHeader(http.Header{Name: "Content-Type", Value: "text/plain"})
		return nil
	}))
	if err != nil {
		t.Fatalf("failed setting up handler: %v", err)
	}
	ctx, cfunc := context.WithCancel(context.Background())
	servClosed := make(chan bool, 1)

	// Start server in background
	go func() {
		defer func() { servClosed <- true }()
		err := httpServer.StartServing(ctx)
		if err != nil {
			t.Errorf("failed to serve: %v", err)
			return
		}
	}()
	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:8092")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// First request with keep-alive
	request1 := "GET /test HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"Connection: keep-alive\r\n\r\n"

	// Second request with close to terminate connection
	request2 := "GET /test HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"Connection: close\r\n\r\n"

	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write([]byte(request1))
	if err != nil {
		t.Fatalf("Failed to write first request: %v", err)
	}

	time.Sleep(500 * time.Millisecond) // brief pause

	_, err = conn.Write([]byte(request2))
	if err != nil {
		t.Fatalf("Failed to write second request: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	respBuf := new(strings.Builder)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		respBuf.WriteString(scanner.Text())
		respBuf.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	response := respBuf.String()
	count := strings.Count(response, "HTTP/1.1")
	if count < 2 {
		t.Fatalf("expected 2 HTTP responses, got %d\nFull response:\n%s", count, response)
	}

	t.Logf("✅ Keep-Alive is working correctly: received %d responses", count)
	cfunc()
	_ = <-servClosed
}
