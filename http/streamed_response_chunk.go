package http

type StreamedResponseChunk struct {
	Data []byte
	Err  error
}
