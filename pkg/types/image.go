package types

import "net/url"

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
