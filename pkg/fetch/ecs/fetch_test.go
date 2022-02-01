package ecs

import (
	"testing"

	"github.com/hown3d/kevo/pkg/fetch/mocks"
	"github.com/hown3d/kevo/pkg/log"
)

func newTestFetcher(t *testing.T, ecs *mocks.ECSAPI, secretsManager *mocks.SecretsManagerAPI) fetcher {
	return fetcher{
		logger:         log.TestLogger{T: t},
		ecs:            ecs,
		secretsmanager: secretsManager,
	}
}
