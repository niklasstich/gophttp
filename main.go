package main

import (
	"context"
	"errors"
	"fmt"
	"http/common"
	"http/handlers"
	"http/http"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

var routes *common.RadixTree[handlers.Handler]

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancel()
			fmt.Println("received SIGINT")
		}
	}()

	//build radix tree of current working directory and register a file handler for every file we find
	routes = common.NewRadixTree[handlers.Handler]()
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	files, err := common.ListFilesRecursive(pwd)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		err = addFileRoute(routes, file)
		if err != nil {
			panic(err)
		}
	}

	sock, err := net.Listen("tcp", ":4488")
	tcpSock := sock.(*net.TCPListener)
	if err != nil {
		panic(err)
	}

	defer sock.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down")
			return
		default:
			// continue
			connectLoop(tcpSock)
		}
	}
}

func addFileRoute(routes *common.RadixTree[handlers.Handler], file string) error {
	parts := strings.Split(file, string(os.PathSeparator))
	path := strings.Join(parts, "/")
	path = fmt.Sprintf("/%s/", path)
	err := routes.Insert(path, handlers.NewFileHandler(file))
	return err
}

func connectLoop(tcpSock *net.TCPListener) {
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
	go handleConnection(conn)
}

func handleConnection(conn net.Conn) {
	//defer closing connection
	resp := &http.Response{}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(conn)
	//queue writing response to connection (we must always answer with at least something, no matter how hard we error out)
	defer writeResponseToConn(resp, conn)

	//parse the request
	req, err := http.ParseRequest(conn)
	if err != nil {
		if errors.Is(err, http.ErrInvalidRequest) {
			err := handlers.BadRequestHandler(req, resp)
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
	fmt.Printf("%+v\n", req)

	if !strings.HasSuffix(req.Path, "/") {
		req.Path += "/"
	}
	route, err := routes.Find(req.Path)
	if err != nil {
		if errors.Is(err, common.ErrNoMatch) {
			//this handler never errors
			_ = handlers.NotFoundHandler(req, resp)
			return
		} else {
			err := fmt.Errorf("error fetching handler from radix tree: %w", err)
			if err != nil {
				panic(err)
			}
			_ = handlers.InternalServerErrorHandler(req, resp)
			return
		}

	} else {
		//call route handler
		err = route.Data.HandleRequest(req, resp)
		if err != nil {
			err := fmt.Errorf("error in handler: %w", err)
			if err != nil {
				panic(err)
			}
			_ = handlers.InternalServerErrorHandler(req, resp)
		}
	}
}

func writeResponseToConn(resp *http.Response, conn net.Conn) {
	err := resp.WriteToConn(conn)
	if err == nil {
		return
	}
	err = fmt.Errorf("error writing response to conn: %w", err)
	if err != nil {
		panic(err)
	}
}
