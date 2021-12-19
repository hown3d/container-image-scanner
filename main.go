package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/ecs"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/kubernetes"
)

func main() {
	for name, fetcher := range fetch.Fetchers {
		fmt.Printf("Fetching images from %v", name)
		images, err := fetcher.GetImages(context.Background())
		if err != nil {
			log.Fatalf("Error!")
		}
		fmt.Println(images)
	}
}
