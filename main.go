package main

import (
	"context"
	"errors"
	"fmt"
	"http/http"
	"net"
	"os"
	"os/signal"
	"time"
)

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
	go topHandler(conn)
}

func topHandler(conn net.Conn) {

	//parse the request
	req, err := http.ParseRequest(conn)
	if err != nil {
		panic(err)
	}
	//print the request
	fmt.Printf("%+v\n", req)

	//send empty response
	write, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	fmt.Println(write)
	if err != nil {
		fmt.Println(err)
		return
	}
	//close connection
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(conn)
}
