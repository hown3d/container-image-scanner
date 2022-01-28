package ecs

import (
	"testing"

	"github.com/hown3d/kevo/mocks"
)

func newTestFetcher(t *testing.T, ecs *mocks.ECSAPI, secretsManager *mocks.SecretsManagerAPI) ecsFetcher {
	return ecsFetcher{
		secretsmanager: secretsManager,
		ecs:            ecs,
	}
}
