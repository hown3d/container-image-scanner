package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/ecs"

	_ "github.com/hown3d/container-image-scanner/pkg/fetch/kubernetes"
	"github.com/hown3d/container-image-scanner/pkg/scan/trivy"
	"github.com/hown3d/container-image-scanner/pkg/types"
)

var (
	trivyURL *string = flag.String("trivyServer", "", "URL of the trivy server")
)

func main() {
	flag.Parse()
	var images []types.Image
	trivy := trivy.New(*trivyURL)

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(len(fetch.Fetchers))
	for name, fetcher := range fetch.Fetchers {
		go func(name string, f fetch.Fetcher) {
			defer wg.Done()
			fmt.Printf("Fetching images from %v\n", name)
			i, err := f.GetImages(context.Background())
			mu.Lock()
			defer mu.Unlock()
			images = append(images, i...)
			if err != nil {
				log.Fatal(err)
			}
		}(name, fetcher)
	}
	wg.Wait()

	for _, image := range images {
		wg.Add(1)
		go func(image types.Image) {
			defer wg.Done()
			fmt.Printf("Scanning image=%v \n", image.String())
			vulnerabilities, err := trivy.Scan(image)
			if err != nil {
				log.Fatal(err)
			}

			for _, v := range vulnerabilities {
				log.Printf("Level=%v Package=%v InstalledVersion=%v\n", v.Level, v.Package, v.CurrentVersion)
			}

		}(image)
	}
	wg.Wait()
}
