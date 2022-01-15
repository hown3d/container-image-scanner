package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hown3d/kevo/pkg/scan"
	"github.com/hown3d/kevo/pkg/scan/trivy"
	"github.com/hown3d/kevo/pkg/types"
	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
	"github.com/sirupsen/logrus"
)

type Kevo struct {
	scanner scan.Scanner
}

var _ kevopb.KevoServiceServer = (*Kevo)(nil)

func NewKevo(scanServerAddress string) Kevo {
	scanner := trivy.New(scanServerAddress)
	return Kevo{
		scanner: scanner,
	}
}

func (k Kevo) SendImage(ctx context.Context, req *kevopb.SendImageRequest) (*kevopb.SendImageResponse, error) {
	image := types.ProtoToInternalImage(req)
	logrus.Infof("Scanning Image %v", image)
	vulnerabilities, err := k.scanner.Scan(image)
	if err != nil {
		err := fmt.Errorf("Failed to scan image %v: %v", image, err)
		log.Println(err)
		return nil, err
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
	return &kevopb.SendImageResponse{}, nil
}
