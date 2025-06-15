package server

import (
	"context"
	"errors"
	"fmt"
	"gophttp/common"
	"gophttp/handlers"
	"gophttp/http"
	"log/slog"
	"math"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

type HttpServer struct {
	routes     *common.RadixTree[RouteHandlerCollection]
	port       int
	reqIndex   uint64
	muReqIndex sync.Mutex
}

var compressionHandler = handlers.NewCompressionHandler()

func NewHttpServer(port int) *HttpServer {
	return &HttpServer{
		routes:   common.NewRadixTree[RouteHandlerCollection](),
		port:     port,
		reqIndex: math.MaxUint64,
	}
}

func (s *HttpServer) nextReqIndex() uint64 {
	s.muReqIndex.Lock()
	defer s.muReqIndex.Unlock()
	s.reqIndex++
	return s.reqIndex
}

// AddRoutes searches for all files and directories under path and adds a handler for each of them to the server
func (s *HttpServer) AddRoutes(path string) error {
	files, err := common.ListFilesRecursive(path)
	if err != nil {
		panic(err)
	}
	dirs, err := common.ListDirsRecursive(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		err = s.addFileRoute(file)
		if err != nil {
			return err
		}
	}

	for _, dir := range dirs {
		err = s.addDirRoute(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HttpServer) insertRoute(route string, method http.Method, handler handlers.Handler) error {
	n, err := s.routes.Find(route)
	if err != nil {
		if errors.Is(err, common.ErrNoMatch) {
			n = NewRouteHandlers()
		} else {
			return err
		}
	}
	n.InsertRoute(method, handler)
	err = s.routes.Insert(route, n)
	return err
}

func (s *HttpServer) addFileRoute(file string) error {
	path := http.GetHttpPathForFilepath(file)
	fh := handlers.NewFileHandler(file)
	h := handlers.ComposeHandlers(fh, compressionHandler)
	err := s.insertRoute(path, http.GET, h)
	return err
}

func (s *HttpServer) addDirRoute(dir string) error {
	path := http.GetHttpPathForFilepath(dir)
	handler, err := handlers.NewDirectoryHandler(dir)
	if err != nil {
		return err
	}
	err = s.insertRoute(path, http.GET, handler)
	return err
}

func (s *HttpServer) StartServing(ctx context.Context) error {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	tcpSock := sock.(*net.TCPListener)
	if err != nil {
		return err
	}

	defer func(sock net.Listener) {
		err := sock.Close()
		if err != nil {
			slog.Error("error closing socket", err)
		}
	}(sock)

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return nil
		default:
			// continue
			s.connectLoop(tcpSock)
		}
	}
}

func (s *HttpServer) connectLoop(tcpSock *net.TCPListener) {
	err := tcpSock.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		slog.Error("error setting socket deadline", err)
		return
	}
	conn, err := tcpSock.Accept()
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			//ignore timeout errors as they are expected
			return
		}
		slog.Error("failed accepting tcp socket connection", err)
		return
	}
	go s.handleConnection(conn)
}

func (s *HttpServer) handleConnection(conn net.Conn) {
	//defer closing connection
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("failed closing socket", err)
		}
	}(conn)

	//handle keep-alive: only exit this loop whenever we get Connection: close, error out, time out or HTTP/1.0
	for {
		shouldClose := s.handleTCPMessage(conn)
		if shouldClose {
			break
		}
	}
}

func (s *HttpServer) handleTCPMessage(conn net.Conn) bool {
	idx := s.nextReqIndex()
	//create an HTTP context with an empty response for the connection
	ctx := http.NewContext(conn, idx)

	//parse the request
	var err error
	ctx.Request, err = http.ParseRequest(ctx)
	//queue writing response to connection (we must always answer with at least something, no matter how hard we error out)
	defer writeResponseToConn(ctx, 0)

	//add common headers required on every response
	defer func(ctx http.Context) {
		err := handlers.ResponseHeadersHandler(ctx)
		if err != nil {
			//handle gracefully? should never error out though
			panic(err)
		}
	}(ctx)
	if err != nil {
		if errors.Is(err, http.ErrInvalidRequest) ||
			errors.Is(err, http.ErrInvalidHttpMethod) ||
			errors.Is(err, http.ErrInvalidHttpVersion) {
			slog.Debug("request failed parsing", "err", err, "index", ctx.Index)
			err := handlers.BadRequestHandler(ctx)
			//bad request handler may never error out
			if err != nil {
				//we messed up big time if we ever get here, error handlers must be error free
				panic(err)
			}
			return true
		}
		panic(err)
	}
	//print the request for debugging
	ra := slog.Group("request",
		"method", ctx.Request.Method,
		"path", ctx.Request.Path,
		"version", ctx.Version,
		"headers", ctx.Request.Headers)
	slog.Debug(ra.String(), "index", ctx.Index)

	routes, err := s.routes.Find(ctx.Request.Path)
	if err != nil {
		if errors.Is(err, common.ErrNoMatch) {
			//this handler never errors
			_ = handlers.NotFoundHandler(ctx)
			return true
		} else {
			err := fmt.Errorf("error fetching handler from radix tree: %w", err)
			if err != nil {
				panic(err)
			}
			_ = handlers.InternalServerErrorHandler(ctx)
			return true
		}
	}
	//try to find handler for HTTP method
	handler := routes.GetRoute(ctx.Request.Method)
	if handler == nil {
		_ = handlers.NotFoundHandler(ctx)
		return true
	}
	err = handler.HandleRequest(ctx)
	if err != nil {
		slog.Error("error in handler", "handler", handler, "err", err, "index", ctx.Index)
		_ = handlers.InternalServerErrorHandler(ctx)
	}

	switch ctx.Request.Version {
	case http.HTTP1_0:
		h, ok := ctx.Request.Headers["Connection"]
		return !(ok && strings.TrimSpace(h.Value) == "keep-alive")
	case http.HTTP1_1:
		h, ok := ctx.Request.Headers["Connection"]
		return ok && strings.TrimSpace(h.Value) == "close"
	default:
		panic("unsupported http version")
	}
}

func writeResponseToConn(ctx http.Context, depth int) {
	if depth == 5 {
		panic("detected recursive loop in writeResponseToConn")
	}
	err := ctx.Response.WriteToConn(ctx.Conn)
	if err == nil {
		return
	}
	if errors.Is(err, net.ErrClosed) {
		slog.Error("unexpected closed connection", "err", err, "index", ctx.Index)
		return
	}
	if errors.Is(err, syscall.EPIPE) {
		slog.Error("unexpected broken pipe", "err", err, "index", ctx.Index)
		return
	}
	if errors.Is(err, syscall.ECONNRESET) {
		slog.Error("unexpected connection reset", "err", err, "index", ctx.Index)
		return
	}
	if errors.Is(err, http.ErrUnknownBodyType) {
		slog.Error("unexpected body type", "err", err, "index", ctx.Index)
		if depth == 0 {
			err = handlers.InternalServerErrorHandler(ctx)
		}
		writeResponseToConn(ctx, depth+1)
		return
	}
	panic(err)
}
