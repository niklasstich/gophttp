package http

import (
	"strconv"
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

func ParseAcceptedQValues(s string) (map[string]float64, error) {
	retval := make(map[string]float64)
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.Split(part, ";")
		if len(kv) == 1 {
			retval[kv[0]] = 1.0
		} else if len(kv) == 2 {
			//parse q value
			s = strings.TrimSpace(kv[1])
			s = strings.TrimPrefix(s, "q=")
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			retval[kv[0]] = f
		} else {
			panic("error reading accepted values in header")
		}
	}
	return retval, nil
}
