package fetch

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hown3d/kevo/ecsAgent/fetch/mocks"
	"github.com/hown3d/kevo/pkg/testutil"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_fetcher_getImagePullSecret(t *testing.T) {
	tests := []struct {
		name          string
		expected      types.RegistryAuth
		secretArn     *string
		mockSetupFunc func(*string, *mocks.SecretsManagerAPI)
		wantErr       bool
	}{
		{
			name:      "existing secret with username and password",
			secretArn: aws.String("test-secret"),
			expected: types.RegistryAuth{
				Username: "test-user",
				Password: "test-pass",
			},
			mockSetupFunc: func(secretArn *string, mockObj *mocks.SecretsManagerAPI) {
				mockObj.On("GetSecretValue",
					mock.Anything,
					&secretsmanager.GetSecretValueInput{
						SecretId: secretArn,
					}).Return(&secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(testutil.GenerateTestRegistryJSON(false, "", "test-user", "test-pass")),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:      "existing secret with docker auth",
			secretArn: aws.String("test-secret"),
			expected: types.RegistryAuth{
				Domain:   "test-domain.com",
				Username: "test-user",
				Password: "test-pass",
			},
			mockSetupFunc: func(secretArn *string, mockObj *mocks.SecretsManagerAPI) {
				mockObj.On("GetSecretValue",
					mock.Anything,
					&secretsmanager.GetSecretValueInput{
						SecretId: secretArn,
					}).
					Return(&secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(testutil.GenerateTestRegistryJSON(true, "test-domain.com", "test-user", "test-pass")),
					}, nil)
			},
			wantErr: false,
		},
		{
			name:      "non existing secret",
			secretArn: aws.String("test-secret"),
			mockSetupFunc: func(secretArn *string, mockObj *mocks.SecretsManagerAPI) {
				mockObj.On("GetSecretValue",
					mock.Anything,
					&secretsmanager.GetSecretValueInput{
						SecretId: secretArn,
					}).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name:      "empty secret arn",
			secretArn: nil,
			expected:  types.RegistryAuth{},
			mockSetupFunc: func(secretArn *string, mockObj *mocks.SecretsManagerAPI) {
				return
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObj := new(mocks.SecretsManagerAPI)
			e := newTestFetcher(t, new(mocks.ECSAPI), mockObj)
			image := &types.Image{}
			tt.mockSetupFunc(tt.secretArn, mockObj)

			err := e.getImagePullSecret(context.Background(), image, tt.secretArn)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, image.Auth, tt.expected)
		})
	}
}
