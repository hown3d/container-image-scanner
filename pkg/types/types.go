package types

import (
	"encoding/json"
	"net/url"
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

type RegistryAuth struct {
	Domain   string
	Username string
	Password string
	// Token to provide for the registry. Will always be
	Token string
}

func (a *RegistryAuth) UnmarshalJSON(data []byte) error {
	var unstructured map[string]interface{}
	err := json.Unmarshal(data, &unstructured)
	if err != nil {
		return err
	}

	// secret can be in two kinds of structures:
	// either
	// {
	//	"auths": {
	//		"docker.accso.de": {
	//			"username": "xyz",
	//			"password": "cyz",
	//			"auth": "xyz"
	//		}
	//	}
	// }
	// or
	// {
	//	"username": "xyz",
	//	"password": "xyz"
	// }

	var registryMap map[string]interface{}
	// check for first kind (docker auth json)
	authMap, ok := unstructured["auths"]
	if ok {
		// check if authMap is in right format
		authMap, ok := authMap.(map[string]interface{})
		if ok {
			// val is the registry map (e.g. docker.accso.de map)
			for key, val := range authMap {
				registryMap, ok = val.(map[string]interface{})
				if ok {
					a.Domain = key
				}
			}
		}
	} else {
		registryMap = unstructured
	}

	pass, passOk := registryMap["password"]
	username, userOk := registryMap["username"]
	if passOk && userOk {
		a.Password = pass.(string)
		a.Username = username.(string)
	}
	return nil
}

type Vulnerability struct {
	Level          string
	Description    string
	Package        string
	CurrentVersion string
	FixedVersion   string
}
