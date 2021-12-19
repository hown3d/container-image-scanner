package util

import "strings"

func SplitImage(image string) (name, tag string) {
	split := strings.Split(image, ":")
	return split[0], split[1]
}
