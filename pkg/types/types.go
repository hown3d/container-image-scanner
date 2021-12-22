package types

type Image struct {
	Name   string
	Tag    string
	Digest string
}

type Vulnerability struct {
	Level          string
	Description    string
	Package        string
	CurrentVersion string
	FixedVersion   string
}
