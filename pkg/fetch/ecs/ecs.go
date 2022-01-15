package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hown3d/kevo/pkg/fetch"
	"github.com/hown3d/kevo/pkg/log"
	"github.com/sirupsen/logrus"
)

const (
	Name        string = "ECS"
	maxServices int    = 10
)

type ecsFetcher struct {
	sess           *session.Session
	ecs            *ecs.ECS
	secretsmanager *secretsmanager.SecretsManager
	logger         log.Logger
}

func init() {
	fetch.Register(Name, newFetcher)
}

func newSession() (*session.Session, error) {
	return session.NewSession(
		&aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	)
}

func newEcsService(sess *session.Session) *ecs.ECS {
	return ecs.New(sess)
}

func newSecretsManagerService(sess *session.Session) *secretsmanager.SecretsManager {
	return secretsmanager.New(sess)
}

func newFetcher() (fetch.Fetcher, error) {
	sess, err := newSession()
	if err != nil {
		return ecsFetcher{}, err
	}

	return ecsFetcher{
		ecs:            newEcsService(sess),
		secretsmanager: newSecretsManagerService(sess),
		logger:         logrus.WithField("fetcher", Name),
	}, nil
}
