package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/hown3d/kevo/pkg/grpc/api"
	tlsConf "github.com/hown3d/kevo/pkg/tls"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	port     *int    = flag.Int("port", 10000, "port to listen on")
	certFile *string = flag.String("cert-file", "", "Path to ssl certificate")
	keyFile  *string = flag.String("key-file", "", "Path to ssl key")
	trivyURL *string = flag.String("trivy-server-url", "", "URL of the trivy server")
	useTls   *bool   = flag.Bool("tls", false, "enable gRPC over TLS")
)

func main() {
	logger := logrus.New()
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		logger.Fatalf("Failed to listen on port %v: %v", *port, err)
	}

	// test trivyURL
	_, err = url.Parse(*trivyURL)
	if err != nil {
		logger.Fatalf("url in trivy-server-url flag is not valid: %v", err)
	} else if *trivyURL == "" {
		logger.Fatal("trivy-server-url flag empty")
	}

	var grpcOpts []grpc.ServerOption
	if *useTls {
		tlsCredentials, err := tlsConf.LoadServerTLSCredentials(*certFile, *keyFile)
		if err != nil {
			logger.Fatalf("cannot load TLS credentials: %v", err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(grpcOpts...)
	api := api.NewKevo(*trivyURL, logger)
	kevopb.RegisterKevoServiceServer(grpcServer, api)
	log.Printf("serving on %v", lis.Addr().String())
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
