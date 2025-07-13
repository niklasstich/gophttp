package http

import (
	"sort"
)

type Header struct {
	Name  string
	Value string
}

type Headers map[string]Header

func (m Headers) HasHeader(key string) bool {
	for _, header := range m {
		if header.Name == key {
			return true
		}
	}
	return false
}

// Sorted returns Headers as a slice of Header, where the headers are sorted alphanumerically ascending by their Name property
func (m Headers) Sorted() []Header {
	headers := make([]Header, 0, len(m))
	for _, header := range m {
		headers = append(headers, header)
	}
	// Sort by Name ascending
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Name < headers[j].Name
	})
	return headers
}
