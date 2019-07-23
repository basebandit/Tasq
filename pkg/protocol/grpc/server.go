package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/basebandit/go-grpc/pkg/api/v1"
	"google.golang.org/grpc"
)

//RunServer runs gRPC service to publish ToDo service
func RunServer(ctx context.Context, v1API v1.ToDoServiceServer, port string) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	//register service
	server := grpc.NewServer()
	v1.RegisterToDoServiceServer(server, v1API)

	//graceful shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			//sig is ^C, handle it
			log.Println("shutting down gRPC server...")

			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	//start gRPC server

	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
