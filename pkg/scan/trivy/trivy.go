package trivy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aquasecurity/fanal/analyzer/config"
	"github.com/aquasecurity/fanal/artifact"
	artifactImage "github.com/aquasecurity/fanal/artifact/image"
	fanalCache "github.com/aquasecurity/fanal/cache"
	"github.com/aquasecurity/fanal/image"
	fanalLog "github.com/aquasecurity/fanal/log"
	fanalTypes "github.com/aquasecurity/fanal/types"
	"github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	"github.com/aquasecurity/trivy/pkg/scanner"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	"github.com/hown3d/container-image-scanner/pkg/log"
	"github.com/hown3d/container-image-scanner/pkg/types"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

type Trivy struct {
	timeout time.Duration
	headers http.Header
	url     string
	logger  log.Logger
	cache   fanalCache.ArtifactCache
	scanner client.Scanner
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

func WithLogger(l log.Logger) func(*Trivy) {
	return func(t *Trivy) {
		t.logger = l
	}
}

func New(url string, options ...func(*Trivy)) Trivy {
	t := Trivy{
		timeout: 1 * time.Second,
		url:     url,
		headers: make(http.Header),
		logger:  logrus.WithField("scanner", "trivy"),
	}
	for _, opt := range options {
		opt(&t)
	}

	remoteScanner := client.NewProtobufClient(client.RemoteURL(url))
	t.scanner = client.NewScanner(client.CustomHeaders(t.headers), remoteScanner)
	t.cache = cache.NewRemoteCache(cache.RemoteURL(t.url), t.headers)

	setupFanalLogger()

	return t
}

func (t Trivy) Scan(image types.Image) (vulnerabilities []types.Vulnerability, err error) {
	//t.logger.Infof("Scanning image %v", image)
	ctx := context.Background()
	sc, cleanUp, err := t.initializeDockerScanner(ctx, image)
	if err != nil {
		return vulnerabilities, err
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
		return vulnerabilities, err
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
		return scanner.Scanner{}, nil, err
	}

	artifact, err := artifactImage.NewArtifact(dockerImage, t.cache, artifact.Option{Quiet: true}, config.ScannerOption{})
	if err != nil {
		return scanner.Scanner{}, nil, err
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
