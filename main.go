package main

import (
	"context"
	"gophttp/server"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cancel()
			slog.Info("received SIGINT")
		}
	}()

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//instantiate server
	serv := server.NewHttpServer(4488)
	err = serv.AddFileRoutes(pwd)
	if err != nil {
		panic(err)
	}
	err = serv.StartServing(ctx)
	if err != nil {
		slog.Error("error in server thread", err)
	}
}
