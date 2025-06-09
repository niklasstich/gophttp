package http

import (
	"strings"
)

func GetHttpPathForFilepath(filepath string) string {
	fp := strings.ReplaceAll(filepath, "\\", "/")
	fp = strings.TrimSpace(fp)
	//handle windows path shenanigans
	if strings.Contains(fp, ":") {
		fp = fp[2:]
	}
	// Normalize root cases
	if fp == "." || fp == "./" || fp == "/." || fp == "/./" {
		return "/"
	}
	// Remove leading "./" or "/."
	fp = strings.TrimPrefix(fp, "./")
	fp = strings.TrimPrefix(fp, "/.")
	// Remove leading and suffix slashes
	fp = strings.Trim(fp, "/")
	return "/" + fp
}
