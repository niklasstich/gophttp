package server

import (
	"context"
	"errors"
	"fmt"
	"gophttp/common"
	"gophttp/handlers"
	"gophttp/http"
	"net"
	"strings"
	"syscall"
	"time"
)

type HttpServer struct {
	Routes *common.RadixTree[handlers.Handler]
	Port   int
}

func NewHttpServer(port int) *HttpServer {
	return &HttpServer{common.NewRadixTree[handlers.Handler](), port}
}

// AddRoutes searches for all files and directories under path and adds a handler for each of them to the server
func (s HttpServer) AddRoutes(path string) error {
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

func (s HttpServer) addFileRoute(file string) error {
	path := http.GetHttpPathForFilepath(file)
	err := s.Routes.Insert(path, handlers.NewFileHandler(file))
	return err
}

func (s HttpServer) addDirRoute(dir string) error {
	path := http.GetHttpPathForFilepath(dir)
	handler, err := handlers.NewDirectoryHandler(dir)
	if err != nil {
		return err
	}
	err = s.Routes.Insert(path, handler)
	return err
}

func (s HttpServer) StartServing(ctx context.Context) error {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	tcpSock := sock.(*net.TCPListener)
	if err != nil {
		return err
	}

	defer func(sock net.Listener) {
		err := sock.Close()
		if err != nil {
			fmt.Printf("error while closing socket: %v", err)
		}
	}(sock)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down")
			return nil
		default:
			// continue
			s.connectLoop(tcpSock)
		}
	}
}

func (s HttpServer) connectLoop(tcpSock *net.TCPListener) {
	err := tcpSock.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := tcpSock.Accept()
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			//ignore timeout errors as they are expected
			return
		}
		fmt.Println(err)
		return
	}
	go s.handleConnection(conn)
}

func (s HttpServer) handleConnection(conn net.Conn) {
	//defer closing connection
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(conn)

	//create an HTTP context with an empty response for the connection
	ctx := http.NewContext(conn)
	ctx.Response = http.NewResponse()

	//parse the request
	var err error
	ctx.Request, err = http.ParseRequest(ctx)
	//queue writing response to connection (we must always answer with at least something, no matter how hard we error out)
	defer writeResponseToConn(ctx.Response, ctx.Conn)

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
			err := handlers.BadRequestHandler(ctx)
			//bad request handler may never error out
			if err != nil {
				//we messed up big time if we ever get here, error handlers must be error free
				panic(err)
			}
			return
		}
		panic(err)
	}
	//print the request for debugging
	//TODO: turn this into toggleable connection trace logging
	fmt.Printf("%+v\n", ctx.Request)

	if !strings.HasSuffix(ctx.Request.Path, "/") {
		ctx.Request.Path += "/"
	}
	route, err := s.Routes.Find(ctx.Request.Path)
	if err != nil {
		if errors.Is(err, common.ErrNoMatch) {
			//this handler never errors
			_ = handlers.NotFoundHandler(ctx)
			return
		} else {
			err := fmt.Errorf("error fetching handler from radix tree: %w", err)
			if err != nil {
				panic(err)
			}
			_ = handlers.InternalServerErrorHandler(ctx)
			return
		}

	} else {
		//call route handler
		err = route.Data.HandleRequest(ctx)
		if err != nil {
			err := fmt.Errorf("error in handler: %w", err)
			if err != nil {
				panic(err)
			}
			_ = handlers.InternalServerErrorHandler(ctx)
		}
	}
}

func writeResponseToConn(resp *http.Response, conn net.Conn) {
	err := resp.WriteToConn(conn)
	if err == nil {
		return
	}
	if errors.Is(err, net.ErrClosed) {
		fmt.Println("tried writing to closed connection: ", err)
		return
	}
	if errors.Is(err, syscall.EPIPE) {
		fmt.Println("tried writing to broken pipe: ", err)
		return
	}
	if errors.Is(err, syscall.ECONNRESET) {
		fmt.Println("connection reset: ", err)
		return
	}
	panic(err)
}
