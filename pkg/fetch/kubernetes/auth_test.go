package kubernetes

import (
	"context"
	"testing"

	"github.com/hown3d/kevo/internal/testutil"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_kubernetesFetcher_getImagePullSecret(t *testing.T) {
	type args struct {
		secretRefs []corev1.LocalObjectReference
		image      *types.Image
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		expected types.RegistryAuth
		fakeObjs []runtime.Object
	}{
		{
			name: "secret exists and is of type docker config",
			args: args{
				secretRefs: []corev1.LocalObjectReference{{Name: "secret"}},
				image:      &types.Image{Name: "testdomain.com/testimage"},
			},
			wantErr: false,
			fakeObjs: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name: "secret",
					},
					Type: corev1.SecretTypeDockerConfigJson,
					Data: generateFakeSecretData("testdomain.com", "user", "pass"),
				},
			},
			expected: types.RegistryAuth{
				Domain:   "testdomain.com",
				Username: "user",
				Password: "pass",
			},
		},
		{
			name: "secret exists and is not of type docker config",
			args: args{
				secretRefs: []corev1.LocalObjectReference{{Name: "secret"}},
				image:      new(types.Image),
			},
			wantErr: true,
			fakeObjs: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name: "secret",
					},
					Type: "blabla",
				},
			},
		},
		{
			name: "secret doesn't exist",
			args: args{
				secretRefs: []corev1.LocalObjectReference{{Name: "secret"}},
				image:      new(types.Image),
			},
			wantErr:  true,
			fakeObjs: []runtime.Object{},
		},
		{
			name: "secret exists but is not for the domain of the image",
			args: args{
				secretRefs: []corev1.LocalObjectReference{{Name: "secret"}},
				image:      &types.Image{Name: "testdomain.com/testimage"},
			},
			wantErr: true,
			fakeObjs: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name: "secret",
					},
					Type: corev1.SecretTypeDockerConfigJson,
					Data: generateFakeSecretData("notyourdomain.com", "user", "pass"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &kubernetesFetcher{
				client: newFakeClient(tt.fakeObjs),
			}
			err := k.getImagePullSecret(context.Background(), tt.args.image, "", tt.args.secretRefs)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expected, tt.args.image.Auth)
		})
	}
}

func generateFakeSecretData(domain, username, password string) map[string][]byte {
	return map[string][]byte{
		".dockerconfigjson": []byte(testutil.GenerateTestRegistryJSON(true, domain, username, password)),
	}
}
