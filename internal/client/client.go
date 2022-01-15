package client

import (
	"context"
	"fmt"
	"log"

	"github.com/hown3d/kevo/pkg/fetch"
	"github.com/hown3d/kevo/pkg/fetch/ecs"
	"github.com/hown3d/kevo/pkg/fetch/kubernetes"
	tlsConf "github.com/hown3d/kevo/pkg/tls"
	"github.com/hown3d/kevo/pkg/types"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"google.golang.org/grpc"
)

type Kevo struct {
	runtime string
	client  kevopb.KevoServiceClient
	fetcher fetch.Fetcher
}

func NewKevo(runtime string, address string, cacertPath string, tls bool) (Kevo, error) {
	var grpcOpts []grpc.DialOption
	if tls {
		credentials, err := tlsConf.LoadClientTLSCredentials(cacertPath)
		if err != nil {
			return Kevo{}, err
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(address, grpcOpts...)
	if err != nil {
		return Kevo{}, err
	}

	fetcher, err := getFetcher(runtime)
	if err != nil {
		return Kevo{}, err
	}

	return Kevo{
		runtime: runtime,
		fetcher: fetcher,
		client:  kevopb.NewKevoServiceClient(conn),
	}, nil
}

// FetchLoop fetches all images of the given fetcher
func (k Kevo) FetchLoop(ctx context.Context) {
	images := make(chan types.Image)
	errors := make(chan error)

	go k.fetcher.Fetch(ctx, images, errors)
	for {
		select {
		case err := <-errors:
			log.Println(err)
		case image := <-images:
			_, err := k.sendImage(ctx, image)
			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (k Kevo) sendImage(ctx context.Context, image types.Image) (*kevopb.SendImageResponse, error) {
	req := types.InternalImageToProto(k.runtime, image)
	return k.client.SendImage(ctx, req)
}

func getFetcher(runtimeType string) (fetcher fetch.Fetcher, err error) {
	switch runtimeType {
	case kubernetes.Name:
		return kubernetes.NewFetcher()
	case ecs.Name:
		return ecs.NewFetcher()
	default:
		return nil, fmt.Errorf("Type %v is not supported", runtimeType)
	}
}
