package kubernetes

import (
	"context"
	"sync"

	"github.com/hown3d/kevo/pkg/types"
	"github.com/hown3d/kevo/pkg/util/imageutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k kubernetesFetcher) GetImages(ctx context.Context) ([]types.Image, error) {
	var images []types.Image
	// empty namespace, to fetch from all namespaces
	podsClient := k.client.CoreV1().Pods("")
	pods, err := podsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, pod := range pods.Items {
		wg.Add(1)
		go func(pod corev1.Pod) {
			defer wg.Done()
			namespace := pod.Namespace
			pullSecrets := pod.Spec.ImagePullSecrets
			mu.Lock()
			defer mu.Unlock()
			images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.ContainerStatuses)...)
			images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.InitContainerStatuses)...)
			images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.EphemeralContainerStatuses)...)
		}(pod)
	}
	wg.Wait()

	return images, nil
}

func (k kubernetesFetcher) getImagesFromContainerStatus(ctx context.Context, namespace string, imagePullSecrets []corev1.LocalObjectReference, status []corev1.ContainerStatus) []types.Image {
	var images []types.Image
	for _, container := range status {
		name, tag := imageutil.SplitImageFromString(container.Image)
		k.logger.Infof("Adding image %v:%v", name, tag)
		image := types.Image{Name: name, Tag: tag}
		k.getImagePullSecret(ctx, &image, namespace, imagePullSecrets)
		images = append(images, image)
	}
	return images
}
