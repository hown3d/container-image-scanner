package ecs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hown3d/kevo/pkg/types"
)

type SecretsManagerAPI interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func newSecretsManagerClient(cfg aws.Config) *secretsmanager.Client {
	return secretsmanager.NewFromConfig(cfg)
}

func (f fetcher) getImagePullSecret(ctx context.Context, image *types.Image, secretArn *string) error {
	// early return when secretArn is not set
	if secretArn == nil {
		return nil
	}
	out, err := f.secretsmanager.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: secretArn,
	})
	if err != nil {
		return err
	}
	if out == nil {
		return errors.New("get secret Value: output is nil")
	}

	return image.Auth.UnmarshalRegistryAuthJSON([]byte(*out.SecretString))
}
