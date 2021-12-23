package kubernetes

import (
	"log"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesFetcher struct {
	client kubernetes.Interface
}

const (
	name string = "Kubernetes"
)

func init() {
	log.Printf("Initializing %v", name)
	fetch.Register(name, newFetcher)
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

func newFetcher() (fetch.Fetcher, error) {
	client, err := newClientSet()
	if err != nil {
		return kubernetesFetcher{}, err
	}
	k := kubernetesFetcher{
		client: client,
	}
	return k, nil
}
