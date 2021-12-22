package kubernetes

import (
	"context"
	"log"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/util/imageutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	f := func() fetch.Fetcher {
		return newFetcher()
	}
	fetch.Register(name, f)
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

func newFetcher() kubernetesFetcher {
	client, err := newClientSet()
	if err != nil {
		log.Println("Can't register kubernetes provider")
	}
	k := kubernetesFetcher{
		client: client,
	}
	return k
}

func (k kubernetesFetcher) GetImages(ctx context.Context) ([]types.Image, error) {
	var images []types.Image
	// empty namespace, to fetch from all namespaces
	podsClient := k.client.CoreV1().Pods("")
	pods, err := podsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		images = append(images, getImagesFromContainerStatus(pod.Status.ContainerStatuses)...)
		images = append(images, getImagesFromContainerStatus(pod.Status.InitContainerStatuses)...)
		images = append(images, getImagesFromContainerStatus(pod.Status.EphemeralContainerStatuses)...)
	}
	return images, nil
}

func getImagesFromContainerStatus(status []corev1.ContainerStatus) []types.Image {
	var images []types.Image
	for _, container := range status {
		name, tag := imageutil.SplitImageFromString(container.Image)

		images = append(images, types.Image{Name: name, Tag: tag})
	}
	return images
}
