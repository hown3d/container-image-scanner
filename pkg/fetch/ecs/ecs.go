package ecs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hown3d/container-image-scanner/pkg/fetch"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/util/imageutil"
)

const (
	name string = "ECS"
)

type ecsFetcher struct {
	svc *ecs.ECS
}

func init() {
	log.Printf("Initializing %v", name)
	f := func() fetch.Fetcher {
		return newFetcher()
	}
	fetch.Register(name, f)
}

func newService() *ecs.ECS {
	sess := session.Must(session.NewSession(
		&aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	))
	return ecs.New(sess)
}

func newFetcher() ecsFetcher {
	return ecsFetcher{
		svc: newService(),
	}
}

func (e ecsFetcher) GetImages(_ context.Context) ([]types.Image, error) {
	var images []types.Image
	clusters, err := e.svc.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't get clusters: %v", err))
	}

	for _, cluster := range clusters.ClusterArns {
		services, err := e.svc.ListServices(&ecs.ListServicesInput{Cluster: cluster})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Can't list services: %v", err))
		}
		for _, service := range services.ServiceArns {
			out, err := e.svc.DescribeServices(&ecs.DescribeServicesInput{
				Cluster:  cluster,
				Services: []*string{service},
			})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Can't describe services: %v", err))
			}
			for _, output := range out.Services {
				out, err := e.svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{TaskDefinition: output.TaskDefinition})
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Can't describe task %v: %v", output.TaskDefinition, err))
				}
				for _, definition := range out.TaskDefinition.ContainerDefinitions {
					// NAME:TAG
					image := *definition.Image
					name, tag := imageutil.SplitImageFromString(image)
					images = append(images, types.Image{Name: name, Tag: tag})
				}
			}
		}
	}
	return images, nil
}
