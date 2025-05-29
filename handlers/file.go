package handlers

import (
	"gophttp/http"
	"os"
	"strings"
)

type fileHandler struct {
	Filepath string
	MIME     string
}

func NewFileHandler(filepath string) Handler {
	f := &fileHandler{Filepath: filepath}
	splits := strings.Split(strings.TrimRight(filepath, "/\\"), ".")

	//TODO: figure out a way to determine most common MIME types
	//TODO: refactor into mime.go
	switch splits[len(splits)-1] {
	case "txt":
		f.MIME = "text/plain"
	case "mp4":
		f.MIME = "video/mp4"
	default:
		f.MIME = "application/octet-stream"
	}
	return f
}

func (f *fileHandler) HandleRequest(req http.Request, resp *http.Response) error {
	//1. write MIME header
	//2. read file
	//3. write body into response
	resp.Headers = append(resp.Headers, http.Header{Name: "Content-Type", Value: f.MIME})

	file, err := os.ReadFile(f.Filepath)
	if err != nil {
		resp.Status = http.StatusInternalServerError
		return err
	}

	resp.Body = file
	resp.Status = http.StatusOK

	return nil
}
