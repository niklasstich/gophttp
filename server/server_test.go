package server_test

import (
	"bufio"
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
	body1 := make([]byte, len(seg1))
	_, err = reader.Read(body1)
	if err != nil {
		t.Fatalf("failed to read first segment: %v", err)
	}
	// Wait for the second segment (should be delayed)
	body2 := make([]byte, len(seg2))
	_, err = reader.Read(body2)
	if err != nil {
		t.Fatalf("failed to read second segment: %v", err)
	}
	response.Write(body1)
	response.Write(body2)

	body := response.String()

	if !strings.Contains(body, string(seg1)) || !strings.Contains(body, string(seg2)) {
		t.Errorf("expected streamed segments in response, got: %q", body)
	}

	if !strings.Contains(body, "Transfer-Encoding: chunked") {
		t.Error("expected 'Transfer-Encoding: chunked' header but didn't find it")
	}
	cfunc()
	_ = <-servClosed
}

func TestStreamedResponseWithDelayAndBrotliAccept(t *testing.T) {
	port := 8091 // Use a different test port
	httpServer := server.NewHttpServer(port)

	const testPath = "/stream-brotli"
	seg1 := []byte("segment1-")
	seg2 := []byte("segment2!")

	handlerFn := handlers.HandlerFunc(func(ctx http.Context) error {
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
	body1 := make([]byte, len(seg1))
	_, err = reader.Read(body1)
	if err != nil {
		t.Fatalf("failed to read first segment: %v", err)
	}
	// Wait for the second segment (should be delayed)
	body2 := make([]byte, len(seg2))
	_, err = reader.Read(body2)
	if err != nil {
		t.Fatalf("failed to read second segment: %v", err)
	}
	response.Write(body1)
	response.Write(body2)

	body := response.String()

	if !strings.Contains(body, string(seg1)) || !strings.Contains(body, string(seg2)) {
		t.Errorf("expected streamed segments in response, got: %q", body)
	}

	if !strings.Contains(body, "Content-Encoding: br") {
		t.Error("expected 'Content-Encoding: br' header but didn't find it")
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
