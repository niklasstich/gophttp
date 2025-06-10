package http

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
