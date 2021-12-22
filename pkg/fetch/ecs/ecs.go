package ecs

import (
	"context"
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
	name        string = "ECS"
	maxServices int    = 10
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

	clusters, err := e.getAllClusters()
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		serviceArns, err := e.getAllServices(cluster)
		if err != nil {
			return nil, err
		}
		clusterImages, err := e.getAllContainerImages(cluster, serviceArns)
		if err != nil {
			return nil, err
		}
		images = append(images, clusterImages...)
	}
	return images, nil
}

func (e ecsFetcher) getAllClusters() (clusterArns []*string, err error) {
	clusters, err := e.svc.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("Can't get clusters: %v", err)
	}
	clusterArns = clusters.ClusterArns
	return clusterArns, nil
}

func (e ecsFetcher) getAllServices(clusterArn *string) (serviceArns []*string, err error) {
	services, err := e.svc.ListServices(&ecs.ListServicesInput{Cluster: clusterArn})
	if err != nil {
		return nil, fmt.Errorf("Can't list services: %v", err)
	}
	serviceArns = services.ServiceArns
	return serviceArns, nil
}

func (e ecsFetcher) getAllContainerImages(clusterArn *string, serviceArns []*string) ([]types.Image, error) {
	resultChan := make(chan types.Image)
	errorChan := make(chan error, 1)

	containerFunc := func(taskDefintionArn *string, imageChan chan types.Image, errorChan chan error) {
		out, err := e.svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
			TaskDefinition: taskDefintionArn,
		})
		if err != nil {
			errorChan <- fmt.Errorf("Can't describe task %v: %v", taskDefintionArn, err)
		}
		for _, definition := range out.TaskDefinition.ContainerDefinitions {
			// NAME:TAG
			image := *definition.Image
			name, tag := imageutil.SplitImageFromString(image)
			imageChan <- types.Image{Name: name, Tag: tag}
		}
	}

	serviceFunc := func(serviceArns [maxServices]*string, errorChan chan error) error {
		var arns []*string
		copy(arns, serviceArns[:])
		out, err := e.svc.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  clusterArn,
			Services: arns,
		})
		if err != nil {
			errorChan <- fmt.Errorf("Can't describe services: %v", err)
		}
		for _, services := range out.Services {
			log.Printf("Getting containers from %v", services.TaskDefinition)
			go containerFunc(services.TaskDefinition, resultChan, errorChan)
		}
		close(resultChan)
		return nil
	}

	for _, chunk := range splitServiceArnsIntoChunks(serviceArns) {
		go serviceFunc(chunk, errorChan)
	}

	var images []types.Image
	for resultChan != nil {
		select {
		case err := <-errorChan:
			return nil, err
		case result, more := <-resultChan:
			images = append(images, result)
			if !more {
				resultChan = nil
			}
		}
	}
	return images, nil
}

// splitServiceArnsIntoChunks splits the given live into maxService sized chunks
func splitServiceArnsIntoChunks(serviceArns []*string) [][maxServices]*string {
	var chunk [maxServices]*string
	chunks := make([][maxServices]*string, 0, len(serviceArns)/maxServices+1)
	for len(serviceArns) >= maxServices {
		// split into chucks and reassign buf onto left strings
		copy(chunk[:maxServices], serviceArns[:maxServices])
		serviceArns = serviceArns[maxServices:]
		chunks = append(chunks, chunk)
	}
	if len(serviceArns) > 0 {
		var leftArns [maxServices]*string
		copy(leftArns[:maxServices], serviceArns[:])
		chunks = append(chunks, leftArns)
	}
	return chunks
}
