package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	"github.com/hown3d/container-image-scanner/pkg/types"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/ecs"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/kubernetes"
)

func main() {
  var images types.Image
	for name, fetcher := range fetch.Fetchers {
		fmt.Printf("Fetching images from %v", name)
		, err := fetcher.GetImages(context.iBackground())

		if err != nil {
			log.Fatalf("Error!")
		}
		fmt.Println(images)
	}
}
