package ecs

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hown3d/kevo/pkg/types"
)

func (e ecsFetcher) getImagePullSecret(image *types.Image, secretArn *string) error {
	// early return when secretArn is not set
	if secretArn == nil {
		return nil
	}
	out, err := e.secretsmanager.GetSecretValue(&secretsmanager.GetSecretValueInput{
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
