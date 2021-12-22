package trivy

import (
	"context"
	"net/http"
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

type registryAuth struct {
	username string
	password string
	// Bearer token to provide for the registry. Will always be
	token string
}

type Trivy struct {
	timeout time.Duration
	headers http.Header
	url     string
	auth    registryAuth
}

func WithCredentialAuth(username, password string) func(*Trivy) {
	return func(t *Trivy) {
		t.auth.username = username
		t.auth.password = password
	}
}

func WithTokenAuth(token string) func(*Trivy) {
	return func(t *Trivy) {
		t.auth.token = token
	}
}

func WithTimeout(timeout time.Duration) func(*Trivy) {
	return func(t *Trivy) {
		t.timeout = timeout
	}
}

func WithCustomHeaders(headers http.Header) func(*Trivy) {
	return func(t *Trivy) {
		t.headers = headers
	}
}
func New(url string, options ...func(*Trivy)) Trivy {
	t := Trivy{
		timeout: 1 * time.Second,
		url:     url,
		headers: make(http.Header),
	}
	for _, opt := range options {
		opt(&t)
	}
	return t
}

func (t Trivy) Scan(image string) (vulnerabilities []types.Vulnerability, err error) {
	ctx := context.Background()
	sc, cleanUp, err := t.initializeDockerScanner(ctx, image)
	if err != nil {
		return []types.Vulnerability{}, err
	}

	defer cleanUp()

	rep, err := sc.ScanArtifact(ctx, trivyTypes.ScanOptions{
		// list of vulnerability types (os,library)
		VulnType: []string{"os", "library"},
		// list of what security issues to detect
		SecurityChecks:      []string{"vuln"},
		ScanRemovedPackages: true,
		ListAllPackages:     true,
	})
	if err != nil {
		return []types.Vulnerability{}, err
	}

	for _, result := range rep.Results {
		for _, vuln := range result.Vulnerabilities {
			v := types.Vulnerability{
				Level:          vuln.Severity,
				Description:    vuln.Description,
				Package:        vuln.PkgName,
				CurrentVersion: vuln.InstalledVersion,
				FixedVersion:   vuln.FixedVersion,
			}
			vulnerabilities = append(vulnerabilities, v)
		}
	}

	return vulnerabilities, nil
}

func (t Trivy) initializeDockerScanner(ctx context.Context, imageName string) (scanner.Scanner, func(), error) {
	remoteScanner := client.NewProtobufClient(client.RemoteURL(t.url))
	clientScanner := client.NewScanner(client.CustomHeaders(t.headers), remoteScanner)
	artifactCache := cache.NewRemoteCache(cache.RemoteURL(t.url), t.headers)

	dockerOption := fanalTypes.DockerOption{
		UserName:      t.auth.username,
		Password:      t.auth.password,
		RegistryToken: t.auth.token,
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
