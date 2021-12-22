package imageutil

import (
	"strings"

	"github.com/hown3d/container-image-scanner/pkg/types"
)

func SplitImageFromString(image string) (name, tag string) {
	split := strings.Split(image, ":")
	return split[0], split[1]
}

func RestoreImageFromStruct(i types.Image) string {
	name := i.Name + ":" + i.Tag
	if i.Digest != "" {
		name = name + "@" + i.Digest
	}
	return name
}
