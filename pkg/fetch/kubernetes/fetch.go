package kubernetes

import (
	"context"
	"errors"

	"github.com/hown3d/kevo/pkg/types"
	"github.com/hown3d/kevo/pkg/util/imageutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func (k *kubernetesFetcher) Fetch(ctx context.Context, imageChan chan types.Image, errChan chan error) {
	controller := k.newPodController(imageChan, errChan)
	controller.Run(ctx.Done())
}

func (k *kubernetesFetcher) newPodController(imageChan chan types.Image, errChan chan error) cache.Controller {

	watchlist := cache.NewListWatchFromClient(
		k.client.CoreV1().RESTClient(),
		string(corev1.ResourcePods),
		corev1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&corev1.Pod{},
		resyncTime, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				pod, ok := newObj.(*corev1.Pod)
				if !ok {
					// will not likely happen, because we only listen to changes for pods
					errChan <- errors.New("Object is not a pod")
				}
				k.sendPodImagesToChan(context.TODO(), pod, imageChan)
			},
			AddFunc: func(obj interface{}) {
				pod, ok := obj.(*corev1.Pod)
				if !ok {
					// not a pod, so return
					// will not likely happen
					errChan <- errors.New("Object is not a pod")
				}
				k.sendPodImagesToChan(context.TODO(), pod, imageChan)
			},
		})
	return controller
}

func (k *kubernetesFetcher) sendPodImagesToChan(ctx context.Context, pod *corev1.Pod, imageChan chan types.Image) {
	var images []types.Image
	namespace := pod.Namespace
	pullSecrets := pod.Spec.ImagePullSecrets
	images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.ContainerStatuses)...)
	images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.InitContainerStatuses)...)
	images = append(images, k.getImagesFromContainerStatus(ctx, namespace, pullSecrets, pod.Status.EphemeralContainerStatuses)...)
	for _, image := range images {
		imageChan <- image
	}
}

func (k *kubernetesFetcher) getImagesFromContainerStatus(ctx context.Context, namespace string, imagePullSecrets []corev1.LocalObjectReference, status []corev1.ContainerStatus) []types.Image {
	var images []types.Image
	for _, container := range status {
		name, tag := imageutil.SplitImageFromString(container.Image)
		image := types.Image{Name: name, Tag: tag}
		k.getImagePullSecret(ctx, image, namespace, imagePullSecrets)
		images = append(images, image)
	}
	return images
}
