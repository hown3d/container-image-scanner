package client

import (
	"context"
	"fmt"

	"github.com/hown3d/kevo/pkg/fetch"
	tlsConf "github.com/hown3d/kevo/pkg/tls"
	"github.com/hown3d/kevo/pkg/types"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"golang.org/x/sync/errgroup"
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
	}

	conn, err := grpc.Dial(address, grpcOpts)
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

// FetchImages fetches all images of the given fetcher
func (k Kevo) FetchImages(ctx context.Context) error {
	g := new(errgroup.Group)

	images, err := k.fetcher.GetImages(ctx)
	if err != nil {
		return err
	}

	for _, image := range images {
		image := image // https://golang.org/doc/faq#closures_and_goroutines
		// Launch a goroutine to send the image.
		g.Go(func() error {
			_, err := k.sendImage(ctx, image)
			return err
		})
	}
	return g.Wait()
}

func (k Kevo) sendImage(ctx context.Context, image types.Image) (*kevopb.SendImageResponse, error) {
	req := types.InternalImageToProto(k.runtime, image)
	return k.client.SendImage(ctx, req)
}

func getFetcher(runtimeType string) (fetch.Fetcher, error) {
	fetcher, ok := fetch.Fetchers[runtimeType]
	if !ok {
		return nil, fmt.Errorf("Runtime type %v is not a registered fetcher", runtimeType)
	}
	return fetcher, nil
}
