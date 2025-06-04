package http

import (
	"os"
	"strings"
)

func GetHttpPathForFilepath(filepath string) string {
	parts := strings.Split(filepath, string(os.PathSeparator))
	s := "/" + strings.Join(parts, "/") + "/"
	return strings.Replace(s, "/.", "", 1)
}
