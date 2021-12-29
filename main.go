package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	"github.com/hown3d/container-image-scanner/pkg/fetch"
	_ "github.com/hown3d/container-image-scanner/pkg/fetch/ecs"
	"github.com/sirupsen/logrus"

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
				logrus.Fatal(err)
			}
		}(name, fetcher)
	}
	wg.Wait()

	//Fixme: Theres currently a bug when running scan in parallel, that the programm crashes
	for _, image := range images {
		wg.Add(1)
		go func(image types.Image) {
			defer wg.Done()
			logrus.Infof("Scanning Image %v", image)
			vulnerabilities, err := trivy.Scan(image)
			if err != nil {
				logrus.Errorf("Failed to scan image %v: %v", image, err)
				return
			}
			for _, v := range vulnerabilities {
				//fmt.Printf("Image=%v Level=%v Package=%v InstalledVersion=%v FixedVersion=%v\nDescription=%v\n\n\n",
				//image, v.Level, v.Package, v.CurrentVersion, v.FixedVersion, v.Description)
				jsonData, err := json.Marshal(v)
				if err != nil {
					logrus.Fatal(err)
				}
				fmt.Println(string(jsonData))
			}
		}(image)
	}
	wg.Wait()
}
