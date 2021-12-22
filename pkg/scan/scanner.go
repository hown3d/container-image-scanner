package scan

import (
	"github.com/hown3d/container-image-scanner/pkg/types"
)

type Scanner interface {
	// Scan gets passed a imagename
	Scan(string) ([]types.Vulnerability, error)
}
