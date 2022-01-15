package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"google.golang.org/grpc"
)

var (
	port int = *flag.Int("port", 10000, "port to listen on")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	api
	kevopb.RegisterKevoServiceServer(grpcServer)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
