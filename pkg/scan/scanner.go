package scan

import (
	"github.com/hown3d/kevo/pkg/types"
)

type Scanner interface {
	// Scan gets passed a imagename
	Scan(types.Image) ([]types.Vulnerability, error)
}
