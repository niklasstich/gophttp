package handlers

import (
	"gophttp/common"
	"gophttp/http"
	"os"
)

type fileHandler struct {
	Filepath string
	MIME     string
}

func NewFileHandler(filepath string) Handler {
	mime, err := common.GetMIMEFromPath(filepath)
	if err != nil {
		panic(err)
	}
	f := &fileHandler{Filepath: filepath, MIME: mime}
	return f
}

func (f *fileHandler) HandleRequest(ctx http.Context) error {
	//1. write MIME header
	//2. read file
	//3. write body into response
	ctx.Response.AddHeader(http.Header{Name: "Content-Type", Value: f.MIME})

	file, err := os.ReadFile(f.Filepath)
	if err != nil {
		ctx.Response.Status = http.StatusInternalServerError
		return err
	}

	ctx.Response.Body = file
	ctx.Response.Status = http.StatusOK

	return nil
}
