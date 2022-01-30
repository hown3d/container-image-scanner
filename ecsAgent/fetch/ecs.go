package fetch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/hown3d/kevo/pkg/util"
)

type ECSAPI interface {
	DescribeTaskDefinition(ctx context.Context, input *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error)
}

func newEcsAPI(cfg aws.Config) *ecs.Client {
	return ecs.NewFromConfig(cfg)
}

func (f fetcher) GetContainerImages(ctx context.Context, taskDefArn *string) ([]types.Image, error) {
	def, err := f.ecs.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{TaskDefinition: taskDefArn})
	if err != nil {
		return nil, fmt.Errorf("describing task definition: %w", err)
	}

	found := []types.Image{}
	if def.TaskDefinition == nil {
		return found, nil
	}
	for _, container := range def.TaskDefinition.ContainerDefinitions {
		name, tag, digest := util.ParseImageReference(*container.Image)
		image := types.Image{
			Name:   name,
			Tag:    tag,
			Digest: digest,
		}
		if container.RepositoryCredentials != nil {
			err := f.getImagePullSecret(ctx, &image, container.RepositoryCredentials.CredentialsParameter)
			if err != nil {
				return nil, fmt.Errorf("getting image pull secret for %v: %w", container.Name, err)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("sending image %v: %w", image.Name, err)
		}
		found = append(found, image)
	}
	return found, nil
}

func (f fetcher) SendImage(ctx context.Context, image types.Image) {
}
