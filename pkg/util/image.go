package util

import (
	"strings"
)

// ParseImageReference returns the name, tag and digest of an image
// Images in Amazon ECR repositories are specified by either using the full
// registry/repository:tag or registry/repository@digest. For example,
// 012345678910.dkr.ecr..amazonaws.com/:latest or
// 012345678910.dkr.ecr..amazonaws.com/@sha256:94afd1f2e64d908bc90dbca0035a5b567EXAMPLE.
//
// *
// Images in official repositories on Docker Hub use a single name (for example,
// ubuntu or mongo).
//
// * Images in other repositories on Docker Hub are qualified
// with an organization name (for example, amazon/amazon-ecs-agent).
//
// * Images in
// other online repositories are qualified further by a domain name (for example,
// quay.io/assemblyline/ubuntu).
func ParseImageReference(s string) (name, tag, digest string) {
	withTag := strings.Split(s, ":")
	if len(withTag) == 2 {
		return withTag[0], withTag[1], ""
	}
	withDigest := strings.Split(s, "@")
	if len(withDigest) == 2 {
		return withDigest[0], "", withDigest[1]
	}

	return s, "", ""
}
