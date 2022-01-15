package api

import (
	"context"

	kevopb "github.com/hown3d/kevo/proto/kevo/v1alpha1"
)

type Kevo struct {
}

func NewKevo() *Kevo {
	return &Kevo{}
}

func (k Kevo) SendImage(context.Context, *kevopb.SendImageRequest) (*kevopb.SendImageResponse, error) {
	panic("not implemented")
}
