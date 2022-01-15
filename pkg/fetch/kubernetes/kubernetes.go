package kubernetes

import (
	"time"

	"github.com/hown3d/kevo/pkg/fetch"
	"github.com/hown3d/kevo/pkg/log"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesFetcher struct {
	client     kubernetes.Interface
	logger     log.Logger
	controller cache.Controller
}

const (
	Name       string = "kubernetes"
	resyncTime        = time.Second * 30
)

func NewFetcher() (fetch.Fetcher, error) {
	client, err := newClientSet()
	if err != nil {
		return &kubernetesFetcher{}, err
	}

	k := kubernetesFetcher{
		client: client,
		logger: logrus.WithField("fetcher", Name),
	}

	return &k, nil
}

func buildConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return config.ClientConfig()
}

func newClientSet() (kubernetes.Interface, error) {
	c, err := buildConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(c)
}
