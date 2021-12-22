package fetch

import (
  "context"
"github.com/hown3d/container-image-scanner/pkg/types")

type Fetcher interface {
	GetImages(context.Context) ([]types.Image, error)
}


// Register can be called from init() on a plugin in this package
// It will automatically be added to the Fetchers map to be called externally
func Register(name string, f FetcherFactory) {
	fetcher := f()
	Fetchers[name] = fetcher
}

// InputFactory lets us use a closure to get intsances of the plugin struct
type FetcherFactory func() Fetcher

// Inputs registry
var Fetchers = map[string]Fetcher{}
