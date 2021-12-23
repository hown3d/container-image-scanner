package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

func Test_ecsFetcher_getImagePullSecret(t *testing.T) {
	type args struct {
		secretArn *string
	}
	tests := []struct {
		name     string
		args     args
		wantAuth bool
		wantErr  bool
	}{
		{
			name: "existing secret with username and password",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:191630647891:secret:accso-ti-cloudmanager-registry-z5iyOB"),
			},
			wantAuth: true,
			wantErr:  false,
		},
		{
			name: "existing secret with docker auth",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:191630647891:secret:accso-infrastruktur-am-registry-k8s-zXfZYG"),
			},
			wantAuth: true,
			wantErr:  false,
		},
		{
			name: "non existing secret",
			args: args{
				aws.String("arn:aws:secretsmanager:eu-central-1:awdawdbajawdjajwd"),
			},
			wantAuth: false,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestFetcher(t)
			got, err := e.getImagePullSecret(tt.args.secretArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ecsFetcher.getImagePullSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantAuth {
				assert.True(t, assert.NotEmpty(t, got.Username))
				assert.True(t, assert.NotEmpty(t, got.Password))
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
