package main

import (
	"context"
	"fmt"
	"gophttp/server"
	"os"
	"os/signal"
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

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//instantiate server
	serv := server.NewHttpServer(4488)
	err = serv.AddRoutes(pwd)
	if err != nil {
		panic(err)
	}
	err = serv.StartServing(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
