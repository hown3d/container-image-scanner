package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/caarlos0/env/v6"
	"github.com/hown3d/kevo/pkg/fetch/ecs"
	"github.com/hown3d/kevo/pkg/grpc/client"
	"github.com/hown3d/kevo/pkg/tls"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var logger = logrus.New()

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	cfg := config{}
	err := parseConfig(&cfg)
	if err != nil {
		logger.Error(err)
		return err
	}

	grpcClient, err := setupGRPCClient(cfg)
	if err != nil {
		logger.Errorf("setup grpc client: %v", err)
		return err
	}

	task, err := ecs.UnmarshalTask(event.Detail)
	if err != nil {
		logger.Errorf("creating fetcher: %v", err)
		return err
	}

	// only get images if the desired status is RUNNING
	if *task.DesiredStatus != string(types.DesiredStatusRunning) {
		return nil
	}

	f, err := ecs.New(logger, cfg.AWSRegion)
	if err != nil {
		logger.Errorf("creating fetcher: %v", err)
		return err
	}

	images, err := f.GetContainerImages(ctx, task.TaskDefinitionArn)
	if err != nil {
		logger.Errorf("getting container images: %v", err)
		return err
	}

	for _, image := range images {
		_, err := grpcClient.SendImage(ctx, image)
		if err != nil {
			logger.Error(err)
		}
	}
	return nil
}

func setupGRPCClient(cfg config) (client.Client, error) {
	var tlsOpt grpc.DialOption
	if cfg.TLS {
		creds, err := tls.LoadClientTLSCredentials(cfg.CACertPath)
		if err != nil {
			return client.Client{}, fmt.Errorf("getting tls credentials: %w", err)
		}
		tlsOpt = grpc.WithTransportCredentials(creds)
	} else {
		tlsOpt = grpc.WithInsecure()
	}

	return client.New("ecs", cfg.ServerAddress, tlsOpt)
}

type config struct {
	AWSRegion     string `env:"AWS_REGION" envDefault:"eu-central-1"`
	ServerAddress string `env:"KEVO_SERVER_ADDRESS"`
	CACertPath    string `env:"KEVO_SERVER_CACERT_PATH"`
	TLS           bool   `env:"USE_TLS" envDefault:"false"`
}

func parseConfig(cfg *config) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parsing env variables: %w", err)
	}
	return nil
}

func main() {
	lambda.Start(Handler)
}
