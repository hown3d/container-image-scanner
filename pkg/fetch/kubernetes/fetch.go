package kubernetes

import (
	"context"
	"errors"

	"github.com/hown3d/kevo/pkg/types"
	"github.com/hown3d/kevo/pkg/util"
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
	images = append(images, k.getImagesFromContainerSpec(ctx, namespace, pullSecrets, pod.Spec.Containers)...)
	for _, image := range images {
		imageChan <- image
	}
}

func (k *kubernetesFetcher) getImagesFromContainerSpec(ctx context.Context, namespace string, imagePullSecrets []corev1.LocalObjectReference, specs []corev1.Container) []types.Image {
	var images []types.Image
	for _, container := range specs {
		name, tag, digest := util.ParseImageReference(container.Image)
		image := types.Image{Name: name, Tag: tag, Digest: digest}
		k.getImagePullSecret(ctx, &image, namespace, imagePullSecrets)
		images = append(images, image)
	}
	return images
}
