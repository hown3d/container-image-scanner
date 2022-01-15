package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/hown3d/kevo/internal/api"
	"github.com/hown3d/kevo/pkg/tls"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"google.golang.org/grpc"
)

var (
	port     int    = *flag.Int("port", 10000, "port to listen on")
	certFile string = *flag.String("cert-file", "", "Path to ssl certificate")
	keyFile  string = *flag.String("key-file", "", "Path to ssl key")
	trivyURL string = *flag.String("trivy-server-url", "", "URL of the trivy server")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", port, err)
	}

	tlsCredentials, err := tls.LoadServerTLSCredentials(certFile, keyFile)
	if err != nil {
		log.Fatalf("cannot load TLS credentials: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
	)
	api := api.NewKevo(trivyURL)
	kevopb.RegisterKevoServiceServer(grpcServer, api)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
