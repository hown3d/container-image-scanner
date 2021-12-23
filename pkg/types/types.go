package types

type Image struct {
	Name   string
	Tag    string
	Digest string
	Auth RegistryAuth
}

func (i Image) String() string {
	name := i.Name + ":" + i.Tag
	if i.Digest != "" {
		name = name + "@" + i.Digest
	}
	return name
}

type RegistryAuth struct {
	Username string
	Password string
	// Token to provide for the registry. Will always be
	Token string
}
type Vulnerability struct {
	Level          string
	Description    string
	Package        string
	CurrentVersion string
	FixedVersion   string
}
