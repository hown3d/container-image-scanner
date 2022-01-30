package fetch

import (
	"context"
	"fmt"

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

	return fetcher{
		secretsmanager: newSecretsManagerClient(awsCfg),
		ecs:            newEcsAPI(awsCfg),
		logger:         logger,
	}, nil
}
