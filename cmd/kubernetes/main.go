package main

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/hown3d/kevo/pkg/fetch/kubernetes"
	"github.com/hown3d/kevo/pkg/grpc/client"
	"github.com/hown3d/kevo/pkg/tls"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type config struct {
	ServerAddress string `env:"KEVO_SERVER_ADDRESS"`
	CACertPath    string `env:"KEVO_SERVER_CACERT_PATH"`
	TLS           bool   `env:"USE_TLS" defaultEnv:"false"`
	LogLevel      string `env:"LOG_LEVEL" defaultEnv:"info"`
}

var logger = logrus.New()

func main() {
	cfg := config{}
	err := parseConfig(&cfg)
	if err != nil {
		logger.Fatal(err)
	}

	grpcClient, err := setupGRPCClient(cfg)
	if err != nil {
		logger.Fatalf("setup grpc client: %v", err)
	}

	fetcher, err := kubernetes.NewFetcher()
	if err != nil {
		logger.Fatalf("creating new fetcher: %v", err)
	}

	imageChan := make(chan types.Image)
	errChan := make(chan error)
	ctx := context.Background()

	go fetcher.Fetch(ctx, imageChan, errChan)
	for {
		select {
		case img := <-imageChan:
			_, err := grpcClient.SendImage(ctx, img)
			if err != nil {
				logger.Errorf("sending image %v to api: %v", img.Name, err)
			} else {
				logger.Infof("successfully send image %v")
			}
		case err := <-errChan:
			logger.Errorf("error while fetching images: %v", err)
		}
	}
}

func parseConfig(cfg *config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parsing env variables: %v", err)
	}
	return nil
}

func setupGRPCClient(cfg config) (client.Client, error) {
	var tlsOpt grpc.DialOption
	if cfg.TLS {
		creds, err := tls.LoadClientTLSCredentials(cfg.CACertPath)
		if err != nil {
			return client.Client{}, fmt.Errorf("getting tls credentials: %v", err)
		}
		tlsOpt = grpc.WithTransportCredentials(creds)
	} else {
		tlsOpt = grpc.WithInsecure()
	}

	return client.New("kubernetes", cfg.ServerAddress, tlsOpt)
}
