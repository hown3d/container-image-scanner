package ecs

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hown3d/container-image-scanner/pkg/fetch"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/util/imageutil"
	"github.com/pkg/errors"
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

func (e ecsFetcher) GetImages(_ context.Context) (images []types.Image, err error) {
	errorChan := make(chan error)
	resultChan := make(chan types.Image, 1)

	clusters, err := e.getAllClusters()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, cluster := range clusters {
		serviceArns, err := e.getAllServices(cluster)
		if err != nil {
			return nil, err
		}

		// AWS API can handle a maximum off 10 services at the time
		// split the serviceArns into chunks of size 10
		chunks := splitServiceArnsIntoChunks(serviceArns, maxServices)
		size := len(chunks)
		log.Printf("Adding %v to waitgroup", size)
		wg.Add(size)
		for _, chunk := range chunks {
			go e.getAllContainerImages(cluster, chunk, resultChan, errorChan, &wg)
		}
	}
	go func() {
		for {
			i := <-resultChan
			log.Printf("recieved %v", i)
			images = append(images, i)
		}
	}()
	wg.Wait()
	close(errorChan)

	for e := range errorChan {
		if e != nil {
			err = errors.Wrap(err, e.Error())
		}
	}
	return images, err
}

func (e ecsFetcher) getAllClusters() (clusterArns []*string, err error) {
	log.Println("Getting all clusters")
	clusters, err := e.svc.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("Can't get clusters: %v", err)
	}
	clusterArns = clusters.ClusterArns
	return clusterArns, nil
}

func (e ecsFetcher) getAllServices(clusterArn *string) (serviceArns []*string, err error) {
	log.Printf("Getting all services for cluster %v", *clusterArn)
	services, err := e.svc.ListServices(&ecs.ListServicesInput{Cluster: clusterArn})
	if err != nil {
		return nil, fmt.Errorf("Can't list services: %v", err)
	}
	serviceArns = services.ServiceArns
	return serviceArns, nil
}

func (e ecsFetcher) getContainerImageFromTaskDefinition(taskdefinitionArn *string, resultChan chan types.Image, errorChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := e.svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskdefinitionArn,
	})
	if err != nil {
		log.Println(err)
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
		resultChan <- image
		log.Printf("Added image %v:%v to resultChannel", name, tag)
	}
	log.Printf("Container coroutine for %v is done!", *taskdefinitionArn)
}

func (e ecsFetcher) getAllContainerImages(clusterArn *string, serviceArns []*string, resultChan chan types.Image, errorChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(serviceArns) <= 0 {
		log.Println("Length of serviceArns is not greater zero, skipping services")
		return
	}
	out, err := e.svc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  clusterArn,
		Services: serviceArns,
	})
	if err != nil {
		err := fmt.Errorf("Can't describe services: %v", err)
		log.Println(err.Error())
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
		log.Printf("Getting containers from %v", *services.TaskDefinition)
		go e.getContainerImageFromTaskDefinition(services.TaskDefinition, resultChan, errorChan, wg)
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
