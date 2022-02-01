package ecs

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hown3d/kevo/pkg/log"
)

type fetcher struct {
	secretsmanager SecretsManagerAPI
	ecs            ECSAPI
	logger         log.Logger
}

func New(logger log.Logger, region string) (fetcher, error) {

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(region))
	if err != nil {
		return fetcher{}, fmt.Errorf("loading default config: %w", err)
	}

	localstack := os.Getenv("LOCALSTACK_HOSTNAME")
	if localstack != "" {
		logger.Infof("Using localstack endpoint %v", localstack)
		localStackResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: localstack,
			}, nil
		})
		awsCfg.EndpointResolverWithOptions = localStackResolver
	}

	return fetcher{
		secretsmanager: newSecretsManagerClient(awsCfg),
		ecs:            newEcsAPI(awsCfg),
		logger:         logger,
	}, nil
}
