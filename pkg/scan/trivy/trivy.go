package trivy

import (
	"context"
	"fmt"
	"time"

	"github.com/aquasecurity/fanal/analyzer/config"
	"github.com/aquasecurity/fanal/artifact"
	artifactImage "github.com/aquasecurity/fanal/artifact/image"
	"github.com/aquasecurity/fanal/image"
	fanalTypes "github.com/aquasecurity/fanal/types"
	"github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	"github.com/aquasecurity/trivy/pkg/scanner"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/types"
)

type RegistryAuth struct {
	Username string
	Password string
	// Bearer Token to provide for the registry. Will always be
	Token string
}

type Trivy struct {
	timeout time.Duration
	url     string
	auth    RegistryAuth
}

func WithCredentialAuth(username, password string) func(*Trivy) {
	return func(t *Trivy) {
		t.auth.Username = username
		t.auth.Password = password
	}
}

func WithTokenAuth(token string) func(*Trivy) {
	return func(t *Trivy) {
		t.auth.Token = token
	}
}

func WithTimeout(timeout time.Duration) func(*Trivy) {
	return func(t *Trivy) {
		t.timeout = timeout
	}
}

func New(url string, options ...func(*Trivy)) Trivy {
	t := Trivy{
		timeout: 1 * time.Second,
		url:     url,
	}
	for _, opt := range options {
		opt(&t)
	}
	return t
}

func (t Trivy) Scan(image string) ([]types.Vulnerability, error) {
	ctx := context.Background()
	sc, cleanUp, err := t.initializeDockerScanner(ctx, image, client.CustomHeaders{})
	if err != nil {
		return []types.Vulnerability{}, err
	}

	defer cleanUp()

	results, err := sc.ScanArtifact(ctx, trivyTypes.ScanOptions{
		VulnType:            []string{"os", "library"},
		ScanRemovedPackages: true,
		ListAllPackages:     true,
	})
	if err != nil {
		return []types.Vulnerability{}, err
	}

	fmt.Println(results)
	return []types.Vulnerability{}, nil
}

func (t Trivy) initializeDockerScanner(ctx context.Context, imageName string, customHeaders client.CustomHeaders) (scanner.Scanner, func(), error) {
	remoteScanner := client.NewProtobufClient(client.RemoteURL(t.url))
	clientScanner := client.NewScanner(customHeaders, remoteScanner)
	artifactCache := cache.NewRemoteCache(cache.RemoteURL(t.url), nil)

	dockerOption := fanalTypes.DockerOption{
		UserName:      t.auth.Username,
		Password:      t.auth.Password,
		RegistryToken: t.auth.Token,
	}

	dockerImage, cleanup, err := image.NewDockerImage(ctx, imageName, dockerOption)
	if err != nil {
		return scanner.Scanner{}, nil, err
	}

	artifact, err := artifactImage.NewArtifact(dockerImage, artifactCache, artifact.Option{}, config.ScannerOption{})
	if err != nil {
		return scanner.Scanner{}, nil, err
	}

	scanner := scanner.NewScanner(clientScanner, artifact)
	return scanner, func() {
		cleanup()
	}, nil
}
