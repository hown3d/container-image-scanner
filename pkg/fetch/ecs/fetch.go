package ecs

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/hown3d/kevo/pkg/util/imageutil"
	"github.com/pkg/errors"
)

func (e ecsFetcher) Fetch(ctx context.Context, imageChan chan types.Image, errorChan chan error) {
	clusters, err := e.getAllClusters(ctx)
	if err != nil {
		errorChan <- err
		return
	}

	var wg sync.WaitGroup
	for _, cluster := range clusters {
		serviceArns, err := e.getAllServices(ctx, cluster)
		if err != nil {
			errorChan <- err
			return
		}

		// AWS API can handle a maximum off 10 services at the time
		// split the serviceArns into chunks of size 10
		chunks := splitServiceArnsIntoChunks(serviceArns, maxServices)
		size := len(chunks)
		e.logger.Debugf("Adding %v to waitgroup", size)
		wg.Add(size)
		for _, chunk := range chunks {
			go e.getAllContainerImages(ctx, cluster, chunk, imageChan, errorChan, &wg)
		}
	}
	wg.Wait()
	// collector
}

func (e ecsFetcher) getAllClusters(ctx context.Context) (clusterArns []*string, err error) {
	e.logger.Debug("Getting all clusters")
	clusters, err := e.ecs.ListClustersWithContext(ctx, &ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("Can't get clusters: %v", err)
	}
	clusterArns = clusters.ClusterArns
	return clusterArns, nil
}

func (e ecsFetcher) getAllServices(ctx context.Context, clusterArn *string) (serviceArns []*string, err error) {
	e.logger.Debugf("Getting all services for cluster %v", *clusterArn)
	services, err := e.ecs.ListServicesWithContext(ctx, &ecs.ListServicesInput{Cluster: clusterArn})
	if err != nil {
		return nil, fmt.Errorf("Can't list services: %v", err)
	}
	serviceArns = services.ServiceArns
	return serviceArns, nil
}

func (e ecsFetcher) getContainerImageFromTaskDefinition(ctx context.Context, taskDefinitionArn *string, imageChan chan types.Image, errorChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := e.ecs.DescribeTaskDefinitionWithContext(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskDefinitionArn,
	})
	if err != nil {
		e.logger.Errorf("Error getting task definition %v: %v", *taskDefinitionArn, err)
		select {
		case errorChan <- err:
		// we're the first worker to fail
		default:
			e := <-errorChan
			e = errors.Wrap(e, err.Error())
			errorChan <- e
		}
		return
	}

	for _, container := range res.TaskDefinition.ContainerDefinitions {
		name, tag := imageutil.SplitImageFromString(*container.Image)
		image := types.Image{
			Name: name,
			Tag:  tag,
		}

		//check for imagepullsecret
		if container.RepositoryCredentials != nil {
			err := e.getImagePullSecret(&image, container.RepositoryCredentials.CredentialsParameter)
			if err != nil {
				select {
				case errorChan <- err:
				// we're the first worker to fail
				default:
					e := <-errorChan
					e = errors.Wrap(e, err.Error())
					errorChan <- e
				}
				return
			}
		}
		imageChan <- image
		e.logger.Infof("Added image %v", image)
	}
	e.logger.Debugf("Container coroutine for %v is done!", *taskDefinitionArn)
}

func (e ecsFetcher) getAllContainerImages(ctx context.Context, clusterArn *string, serviceArns []*string, imageChan chan types.Image, errorChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(serviceArns) <= 0 {
		e.logger.Debug("Length of serviceArns is not greater zero, skipping services")
		return
	}
	out, err := e.ecs.DescribeServicesWithContext(ctx, &ecs.DescribeServicesInput{
		Cluster:  clusterArn,
		Services: serviceArns,
	})
	if err != nil {
		err := fmt.Errorf("Can't describe services: %v", err)
		e.logger.Error(err.Error())
		select {
		case errorChan <- err:
		// we're the first worker to fail
		default:
			e := <-errorChan
			e = errors.Wrap(e, err.Error())
			errorChan <- e
		}
		return
	}
	wg.Add(len(out.Services))
	for _, services := range out.Services {
		e.logger.Debugf("Getting containers from %v", *services.TaskDefinition)
		go e.getContainerImageFromTaskDefinition(ctx, services.TaskDefinition, imageChan, errorChan, wg)
	}
}

// splitServiceArnsIntoChunks splits the given live into maxService sized chunks
func splitServiceArnsIntoChunks(serviceArns []*string, limit int) [][]*string {
	var chunk []*string
	chunks := make([][]*string, 0, len(serviceArns)/limit+1)
	for len(serviceArns) >= limit {
		// split into chucks and reassign buf onto left strings
		chunk, serviceArns = serviceArns[:limit], serviceArns[limit:]
		chunks = append(chunks, chunk)
	}
	if len(serviceArns) > 0 {
		chunks = append(chunks, serviceArns)
	}
	return chunks
}
