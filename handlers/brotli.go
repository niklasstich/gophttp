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

	if c, ok := ctx.Response.Body.(chan http.StreamedResponseChunk); ok {
		err := b.handleChannel(ctx, c)
		if err != nil {
			return err
		}
	} else { //NOT a channel, do normal stuff
		bbuf, err := castBody(ctx.Response.Body)
		if err != nil {
			return err
		}
		newBuf, err := b.compressBody(bbuf)
		if err != nil {
			ctx.Response.AddHeader(http.Header{
				Name:  "Content-Length",
				Value: strconv.Itoa(len(newBuf)),
			})
			return err
		}
		//assign body to response
		ctx.Response.Body = newBuf
	}

	//set compression header
	ctx.Response.AddHeader(http.Header{
		Name:  "Content-Encoding",
		Value: "br",
	})
	return nil
}

func (b brotliHandler) handleChannel(ctx http.Context, c chan http.StreamedResponseChunk) error {
	//TODO:
	//1. create target channel and set in ctx
	//2. start loop in goroutine:
	//2.1. read from c
	//2.2. get brotli bytes
	//2.3.1 if error, write to err channel
	//2.3.3 otherwise write brotli bytes to target channel
	//3. wait on err channel and return
	tChan := make(chan http.StreamedResponseChunk, 1)
	ctx.Response.Body = tChan
	go func() {
		defer close(tChan)
		for chunk := range c {
			if chunk.Err != nil {
				tChan <- chunk
				return
			}
			compBuf, err := b.compressBody(chunk.Data)
			if err != nil {
				tChan <- http.StreamedResponseChunk{Err: err}
				return
			}
			tChan <- http.StreamedResponseChunk{Data: compBuf}
		}
	}()
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
