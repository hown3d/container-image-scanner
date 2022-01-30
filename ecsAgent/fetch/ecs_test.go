package fetch

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hown3d/kevo/ecsAgent/fetch/mocks"
	"github.com/hown3d/kevo/pkg/testutil"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_fetcher_GetContainerImages(t *testing.T) {
	type args struct {
		taskDefArn *string
	}
	tests := []struct {
		name          string
		args          args
		want          []types.Image
		wantErr       bool
		mockSetupFunc func(secretsManagerMock *mocks.SecretsManagerAPI, ecsMock *mocks.ECSAPI, taskDefArn *string)
	}{
		{
			name: "2 images in task definition",
			mockSetupFunc: func(secretManagerMock *mocks.SecretsManagerAPI, mockObj *mocks.ECSAPI, taskDefArn *string) {
				mockObj.On("DescribeTaskDefinition", mock.Anything, &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefArn,
				}).
					Return(&ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &ecsTypes.TaskDefinition{
							ContainerDefinitions: []ecsTypes.ContainerDefinition{
								ecsTypes.ContainerDefinition{
									Image: aws.String("foo"),
								},
								ecsTypes.ContainerDefinition{
									Image: aws.String("bar"),
								},
							},
						},
					}, nil)
			},
			args: args{
				taskDefArn: aws.String("test"),
			},
			want: []types.Image{
				types.Image{Name: "docker.io/library/foo"},
				types.Image{Name: "docker.io/library/bar"},
			},
			wantErr: false,
		},
		{
			name: "no image in task definition",
			mockSetupFunc: func(secretManagerMock *mocks.SecretsManagerAPI, ecsMock *mocks.ECSAPI, taskDefArn *string) {
				ecsMock.On("DescribeTaskDefinition", mock.Anything, &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefArn,
				}).
					Return(&ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &ecsTypes.TaskDefinition{
							ContainerDefinitions: []ecsTypes.ContainerDefinition{},
						},
					}, nil)
			},
			args: args{
				taskDefArn: aws.String("test"),
			},
			want:    []types.Image{},
			wantErr: false,
		},
		{
			name: "image with registry creds",
			mockSetupFunc: func(secretManagerMock *mocks.SecretsManagerAPI, ecsMock *mocks.ECSAPI, taskDefArn *string) {
				ecsMock.On("DescribeTaskDefinition", mock.Anything, &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefArn,
				}).
					Return(&ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &ecsTypes.TaskDefinition{
							ContainerDefinitions: []ecsTypes.ContainerDefinition{
								{
									Image: aws.String("test-domain.com/foo/bar"),
									RepositoryCredentials: &ecsTypes.RepositoryCredentials{
										CredentialsParameter: aws.String("test-secret-arn"),
									},
								},
							},
						},
					}, nil)

				secretManagerMock.On("GetSecretValue",
					mock.Anything,
					&secretsmanager.GetSecretValueInput{
						SecretId: aws.String("test-secret-arn"),
					}).
					Return(&secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(testutil.GenerateTestRegistryJSON(true, "test-domain.com", "test-user", "test-pass")),
					}, nil)
			},
			args: args{
				taskDefArn: aws.String("test"),
			},
			want: []types.Image{
				{
					Name: "test-domain.com/foo/bar",
					Auth: types.RegistryAuth{
						Domain:   "test-domain.com",
						Username: "test-user",
						Password: "test-pass",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "taskdefinition is nil",
			mockSetupFunc: func(secretManagerMock *mocks.SecretsManagerAPI, ecsMock *mocks.ECSAPI, taskDefArn *string) {
				ecsMock.On("DescribeTaskDefinition", mock.Anything, &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: taskDefArn,
				}).Return(nil, errors.New("error"))

			},
			args: args{
				taskDefArn: aws.String("test"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ecs := new(mocks.ECSAPI)
			sm := new(mocks.SecretsManagerAPI)
			f := newTestFetcher(t, ecs, sm)
			tt.mockSetupFunc(sm, ecs, tt.args.taskDefArn)
			got, err := f.GetContainerImages(context.Background(), tt.args.taskDefArn)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
