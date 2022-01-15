package types

import (
	"net/url"

	kevo "github.com/hown3d/kevo/proto/kevo/v1alpha1"
)

type Image struct {
	Name   string
	Tag    string
	Digest string
	Auth   RegistryAuth
}

func (i Image) String() string {
	name := i.Name + ":" + i.Tag
	if i.Digest != "" {
		name = name + "@" + i.Digest
	}
	return name
}

func (i Image) RegistryDomain() (string, error) {
	// prepend double slashes because url parse needs a scheme
	u, err := url.Parse("//" + i.Name)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

func ProtoToInternalImage(req *kevo.SendImageRequest) Image {
	var image Image
	image.Name = req.Image.Name
	image.Digest = req.Image.Digest
	image.Tag = req.Image.Tag
	image.Auth.Username = req.Auth.Username
	image.Auth.Password = req.Auth.Password
	image.Auth.Token = req.Auth.Token
	return image
}

func InternalImageToProto(runtime string, image Image) *kevo.SendImageRequest {
	return &kevo.SendImageRequest{
		Cluster: runtime,
		Image: &kevo.Image{
			Name:   image.Name,
			Tag:    image.Tag,
			Digest: image.Digest,
		},
		Auth: &kevo.Auth{
			Username: image.Auth.Username,
			Password: image.Auth.Password,
			Token:    image.Auth.Token,
		},
	}
}
