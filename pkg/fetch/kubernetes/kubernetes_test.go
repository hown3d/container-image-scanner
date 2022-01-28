package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func newFakeClient(objects []runtime.Object) kubernetes.Interface {
	return fake.NewSimpleClientset(objects...)
}
