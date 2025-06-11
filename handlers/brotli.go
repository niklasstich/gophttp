package handlers

import (
	"bytes"
	"fmt"
	"github.com/andybalholm/brotli"
	"gophttp/http"
	"strconv"
)

type brotliHandler struct {
	quality int
}

func NewBrotliHandler(quality int) Handler {
	return &brotliHandler{quality}
}

var ErrUnknownBodyType = fmt.Errorf("unknown body type")

func (b brotliHandler) HandleRequest(ctx http.Context) error {
	if !reqAcceptsBrotli(ctx.Request) {
		return nil
	}
	bbuf, err := castBody(ctx.Response.Body)
	if err != nil {
		return err
	}
	newBuf, err := b.compressBody(bbuf)
	if err != nil {
		return err
	}

	//assign body to response and set compression header
	ctx.Response.AddHeader(http.Header{
		Name:  "Content-Encoding",
		Value: "br",
	})
	ctx.Response.AddHeader(http.Header{
		Name:  "Content-Length",
		Value: strconv.Itoa(len(newBuf)),
	})
	ctx.Response.Body = newBuf
	return nil
}

func reqAcceptsBrotli(request *http.Request) bool {
	if !request.Headers.HasHeader("Accept-Encoding") {
		return false
	}
	header := request.Headers["Accept-Encoding"]
	encodings, err := http.ParseAcceptedQValues(header.Value)
	if err != nil {
		return false
	}
	for s := range encodings {
		if s == "br" {
			return true
		}
	}
	return false
}

func (b brotliHandler) compressBody(body []byte) ([]byte, error) {
	var newBuf bytes.Buffer
	//create brotli writer that writes into the response buffer
	writer := brotli.NewWriterOptions(&newBuf, brotli.WriterOptions{
		Quality: b.quality,
		LGWin:   0,
	})
	_, err := writer.Write(body)
	if err != nil {
		return nil, fmt.Errorf("error writing compressed body: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing brotli writer: %v", err)
	}
	return newBuf.Bytes(), nil
}

func castBody(body interface{}) ([]byte, error) {
	var bbuf []byte
	var ok bool
	//copy old buffer
	if bbuf, ok = body.([]byte); !ok {
		if s, ok := body.(string); ok {
			bbuf = []byte(s)
		} else {
			return nil, ErrUnknownBodyType
		}
	}
	return bbuf, nil
}
