package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hown3d/kevo/pkg/fetch"
	"github.com/hown3d/kevo/pkg/types"
	"golang.org/x/sync/errgroup"
)

func getFetcher(runtimeType string) (fetch.Fetcher, error) {
	fetcher, ok := fetch.Fetchers[runtimeType]
	if !ok {
		return nil, fmt.Errorf("Runtime type %v is not a registered fetcher", runtimeType)
	}
	return fetcher, nil
}

func fetchImages(ctx context.Context, fetcher fetch.Fetcher) error {
	g := new(errgroup.Group)

	images, err := fetcher.GetImages(ctx)
	if err != nil {
		return err
	}

	for _, image := range images {
		image := image // https://golang.org/doc/faq#closures_and_goroutines
		// Launch a goroutine to send the image.
		g.Go(func() error {
			return sendImageToServer(image)
		})
	}

	// Wait for all sends to complete.
	return g.Wait()
}

func sendImageToServer(image types.Image) error {
	panic("not implemented")
}

func createAPIDial() {

}

func main() {
	// TODO: load config file
	var runtimeType string

	fetcher, err := getFetcher(runtimeType)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = fetchImages(ctx, fetcher)
	if err != nil {
		log.Fatal(err)
	}
}
