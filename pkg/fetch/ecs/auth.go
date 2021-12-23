package ecs

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hown3d/container-image-scanner/pkg/types"
)

type auth struct {
	username string
	password string
}

func (a *auth) UnmarshalJSON(data []byte) error {
	var unstructured map[string]interface{}
	err := json.Unmarshal(data, &unstructured)
	if err != nil {
		return err
	}

	// secret can be in two kinds of structures:
	// either
	// {
	//	"auths": {
	//		"docker.accso.de": {
	//			"username": "xyz",
	//			"password": "cyz",
	//			"auth": "xyz"
	//		}
	//	}
	// }
	// or
	// {
	//	"username": "xyz",
	//	"password": "xyz"
	// }

	var registryMap map[string]interface{}
	// check for first kind (docker auth json)
	authMap, ok := unstructured["auths"]
	if ok {
		// check if authMap is in right format
		authMap, ok := authMap.(map[string]interface{})
		if ok {
			// val is the registry string (in e.g. docker.accso.de)
			for _, val := range authMap {
				registryMap, ok = val.(map[string]interface{})
			}
		}
	} else {
		registryMap = unstructured
	}

	pass, passOk := registryMap["password"]
	username, userOk := registryMap["username"]
	if passOk && userOk {
		a.password = pass.(string)
		a.username = username.(string)
	}

	return nil
}

func (e ecsFetcher) getImagePullSecret(secretArn *string) (types.RegistryAuth, error) {
	out, err := e.secretsmanager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: secretArn,
	})
	if err != nil {
		return types.RegistryAuth{}, err
	}

	var jsonAuth auth
	secretVal := out.SecretString
	if json.Valid([]byte(*secretVal)) {
		err := json.Unmarshal([]byte(*secretVal), &jsonAuth)
		if err != nil {
			return types.RegistryAuth{}, err
		}
	}
	auth := types.RegistryAuth{
		Username: jsonAuth.username,
		Password: jsonAuth.password,
	}
	return auth, nil
}
