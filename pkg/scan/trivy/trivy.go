package trivy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aquasecurity/fanal/analyzer/config"
	"github.com/aquasecurity/fanal/artifact"
	artifactImage "github.com/aquasecurity/fanal/artifact/image"
	fanalCache "github.com/aquasecurity/fanal/cache"
	"github.com/aquasecurity/fanal/image"
	fanalLog "github.com/aquasecurity/fanal/log"
	fanalTypes "github.com/aquasecurity/fanal/types"
	trivyCache "github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	"github.com/aquasecurity/trivy/pkg/scanner"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	_ "github.com/hashicorp/go-retryablehttp"
	"github.com/hown3d/kevo/pkg/log"
	"github.com/hown3d/kevo/pkg/types"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

type Trivy struct {
	headers    http.Header
	url        string
	logger     log.Logger
	httpClient http.Client
	cache      fanalCache.ArtifactCache
	scanner    client.Scanner
}

type TrivyOption func(*Trivy)

func WithCustomHeaders(headers http.Header) TrivyOption {
	return func(t *Trivy) {
		t.headers = headers
	}
}

func WithLogger(l log.Logger) TrivyOption {
	return func(t *Trivy) {
		t.logger = l
	}
}

func New(url string, options ...func(*Trivy)) Trivy {
	t := Trivy{
		url:     url,
		headers: make(http.Header),
		logger:  logrus.WithField("scanner", "trivy"),
	}
	for _, opt := range options {
		opt(&t)
	}

	setupFanalLogger()
	remoteScanner := client.NewProtobufClient(client.RemoteURL(url))
	t.scanner = client.NewScanner(client.CustomHeaders(t.headers), remoteScanner)
	t.cache = trivyCache.NewRemoteCache(trivyCache.RemoteURL(t.url), t.headers)

	return t
}

func (t Trivy) Scan(image types.Image) (vulnerabilities []types.Vulnerability, err error) {
	//t.logger.Infof("Scanning image %v", image)
	ctx := context.Background()
	sc, cleanUp, err := t.initializeDockerScanner(ctx, image)
	if err != nil {
		return vulnerabilities, fmt.Errorf("initializing docker scanner: %w", err)
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
		return vulnerabilities, fmt.Errorf("scanning artifact: %w", err)
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

func (t Trivy) initializeDockerScanner(ctx context.Context, i types.Image) (scanner.Scanner, func(), error) {
	dockerOption := fanalTypes.DockerOption{
		UserName:      i.Auth.Username,
		Password:      i.Auth.Password,
		RegistryToken: i.Auth.Token,
	}

	dockerImage, cleanup, err := image.NewDockerImage(ctx, i.String(), dockerOption)
	if err != nil {
		return scanner.Scanner{}, nil, fmt.Errorf("creating new docker image: %w", err)
	}

	artifact, err := artifactImage.NewArtifact(dockerImage, t.cache, artifact.Option{Quiet: true}, config.ScannerOption{})
	if err != nil {
		return scanner.Scanner{}, nil, fmt.Errorf("creating new artifact: %w", err)
	}

	scanner := scanner.NewScanner(t.scanner, artifact)
	return scanner, func() {
		cleanup()
	}, nil
}

func setupFanalLogger() error {
	// set logger for fanal
	rawJSON := []byte(`{
  "level": "error",
  "encoding": "json",
  "outputPaths": ["stdout"],
  "errorOutputPaths": ["stderr"],
  "initialFields": {"scanner": "trivy"},
  "encoderConfig": {
  "messageKey": "message",
  "levelKey": "level",
  "levelEncoder": "lowercase"
  }
  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		return err
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}
	fanalLog.SetLogger(zapLogger.Sugar())
	return nil
}
