package ecs

import (
	"encoding/json"

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

	var auth types.RegistryAuth
	secretVal := out.SecretString
	if json.Valid([]byte(*secretVal)) {
		err := json.Unmarshal([]byte(*secretVal), &auth)
		if err != nil {
			return err
		}
		image.Auth = auth
	}
	return nil
}
