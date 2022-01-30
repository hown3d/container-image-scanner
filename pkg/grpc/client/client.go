package client

import (
	"context"

	"github.com/hown3d/kevo/pkg/fetch"
	"github.com/hown3d/kevo/pkg/types"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"google.golang.org/grpc"
)

type Client struct {
	runtime string
	client  kevopb.KevoServiceClient
	fetcher fetch.Fetcher
}

func New(runtime string, serverAddr string, grpcOpts ...grpc.DialOption) (Client, error) {
	conn, err := grpc.Dial(serverAddr, grpcOpts...)
	if err != nil {
		return Client{}, err
	}

	return Client{
		runtime: runtime,
		client:  kevopb.NewKevoServiceClient(conn),
	}, nil
}

// FetchLoop fetches all images of the given fetcher
// func (k Client) FetchLoop(ctx context.Context) {
// 	images := make(chan types.Image)
// 	errors := make(chan error)

// 	go k.fetcher.Fetch(ctx, images, errors)
// 	for {
// 		select {
// 		case err := <-errors:
// 			log.Println(err)
// 		case image := <-images:
// 			_, err := k.sendImage(ctx, image)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }

func (k Client) SendImage(ctx context.Context, image types.Image) (*kevopb.SendImageResponse, error) {
	req := types.InternalImageToProto(k.runtime, image)
	return k.client.SendImage(ctx, req)
}

// func getFetcher(runtimeType string) (fetcher fetch.Fetcher, err error) {
// 	switch runtimeType {
// 	case kubernetes.Name:
// 		return kubernetes.NewFetcher()
// 	case ecs.Name:
// 		return ecs.New()
// 	default:
// 		return nil, fmt.Errorf("Type %v is not supported", runtimeType)
// 	}
// }
