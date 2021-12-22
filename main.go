package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"

	//_ "github.com/hown3d/container-image-scanner/pkg/fetch/ecs"
	"github.com/hown3d/container-image-scanner/pkg/fetch"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/kubernetes"
	"github.com/hown3d/container-image-scanner/pkg/scan/trivy"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/util/imageutil"
)

var (
	trivyURL *string = flag.String("trivyServer", "", "URL of the trivy server")
)

func main() {
	flag.Parse()
	var images []types.Image
	trivy := trivy.New(*trivyURL)
	for name, fetcher := range fetch.Fetchers {
		fmt.Printf("Fetching images from %v\n", name)
		i, err := fetcher.GetImages(context.Background())
		images = append(images, i...)
		if err != nil {
			log.Fatal(err)
		}
	}

	var wg sync.WaitGroup

	for _, image := range images {
		wg.Add(1)
		go func(image types.Image) {
			defer wg.Done()
			i := imageutil.RestoreImageFromStruct(image)
			fmt.Printf("Scanning image=%v \n", i)
			vulnerabilities, err := trivy.Scan(i)
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
