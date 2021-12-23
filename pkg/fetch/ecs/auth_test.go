package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_ecsFetcher_getImagePullSecret(t *testing.T) {
	type args struct {
		secretArn *string
		image     *types.Image
	}
	type expected struct {
		domain string
	}
	tests := []struct {
		name     string
		expected expected
		args     args
		wantAuth bool
		wantErr  bool
	}{
		{
			name: "existing secret with username and password",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:191630647891:secret:accso-ti-cloudmanager-registry-z5iyOB"),
				&types.Image{},
			},
			expected: expected{
				domain: "",
			},
			wantAuth: true,
			wantErr:  false,
		},
		{
			name: "existing secret with docker auth",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:191630647891:secret:accso-infrastruktur-am-registry-k8s-zXfZYG"),
				&types.Image{},
			},
			expected: expected{
				domain: "docker.accso.de",
			},
			wantAuth: true,
			wantErr:  false,
		},
		{
			name: "non existing secret",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:awdawdbajawdjajwd"),
				&types.Image{},
			},
			wantAuth: false,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestFetcher(t)
			err := e.getImagePullSecret(tt.args.image, tt.args.secretArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ecsFetcher.getImagePullSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantAuth {
				assert.True(t, assert.NotEmpty(t, tt.args.image.Auth.Password))
				assert.True(t, assert.NotEmpty(t, tt.args.image.Auth.Username))
				assert.Equal(t, tt.expected.domain, tt.args.image.Auth.Domain)
			}
		})
	}
}

func newTestFetcher(t *testing.T) ecsFetcher {
	sess, err := newSession()
	if err != nil {
		t.Logf("Error creating session: %v", err)
		t.FailNow()
	}
	return ecsFetcher{
		secretsmanager: newSecretsManagerService(sess),
	}
}
