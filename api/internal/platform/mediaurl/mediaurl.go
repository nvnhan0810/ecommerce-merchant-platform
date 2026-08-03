package mediaurl

import "strings"

// Absolute builds a public media proxy URL for an object key.
func Absolute(publicAPIBase, imageKey string) string {
	imageKey = strings.TrimSpace(strings.TrimPrefix(imageKey, "/"))
	if imageKey == "" {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(publicAPIBase), "/")
	if base == "" {
		return "/api/v1/media/" + imageKey
	}
	return base + "/api/v1/media/" + imageKey
}
