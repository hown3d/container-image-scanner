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

func (k Client) SendImage(ctx context.Context, image types.Image) (*kevopb.SendImageResponse, error) {
	req := types.InternalImageToProto(k.runtime, image)
	return k.client.SendImage(ctx, req)
}
