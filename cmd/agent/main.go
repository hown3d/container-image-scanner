package main

import (
	"context"
	"flag"
	"log"

	"github.com/hown3d/kevo/internal/client"
)

var (
	address    string = *flag.String("kevo-api-url", "", "URL of the kevo API server to use")
	cacertFile string = *flag.String("ca-cert-file", "", "Path to CA Certificate, which was used to create kevo api certificate")
)

func main() {
	// TODO: load config file
	var runtimeType string
	kevo, err := client.NewKevo(runtimeType, address, cacertFile)
	if err != nil {
		log.Fatal(err)
	}
	kevo.FetchImages(context.Background())
}
