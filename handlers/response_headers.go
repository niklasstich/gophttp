package handlers

import (
	"fmt"
	"gophttp/common"
	"gophttp/http"
	"strconv"
	"time"
)

const VERSION = "0.1"
const SERVER_NAME = "gophttp"

var ServerHeader = http.Header{
	Name:  "Server",
	Value: fmt.Sprintf("%s/%s", SERVER_NAME, VERSION),
}

var timeFunc = time.Now

func ResponseHeadersHandler(ctx http.Context) error {
	//Server
	ctx.Response.AddHeader(ServerHeader)
	//Date
	ctx.Response.AddHeader(http.Header{
		Name:  "Date",
		Value: common.ToHttpDateFormat(timeFunc()),
	})
	_, bodyIsChannel := ctx.Response.Body.(chan http.StreamedResponseChunk)
	if bodyIsChannel {
		//delete content-length, add transfer-encoding: chunked instead
		delete(ctx.Response.Headers, "Content-Length")
		ctx.Response.AddHeader(http.Header{
			Name:  "Transfer-Encoding",
			Value: "chunked",
		})
	} else {
		if !ctx.Response.Headers.HasHeader("Content-Length") {
			var length int
			switch v := ctx.Response.Body.(type) {
			case string:
				length = len(v)
			case []byte:
				length = len(v)
			default:
				//TODO: figure out what to do
				return nil
			}
			ctx.Response.AddHeader(http.Header{
				Name:  "Content-Length",
				Value: strconv.Itoa(length),
			})
		}
	}

	tryWriteConnectionHeader(ctx)
	return nil
}

func tryWriteConnectionHeader(ctx http.Context) {
	if ctx.Response.Headers.HasHeader("Connection") {
		return
	}
	var connHeader http.Header
	if ctx.Request == nil {
		connHeader = http.Header{
			Name:  "Connection",
			Value: "close",
		}
	} else {
		if h, ok := ctx.Request.Headers["Connection"]; ok {
			connHeader = h
		} else {
			connHeader = http.Header{Name: "Connection"}
			if ctx.Request.Version == http.HTTP1_0 {
				connHeader.Value = "close"
			} else if ctx.Request.Version == http.HTTP1_1 {
				connHeader.Value = "keep-alive"
			} else {
				connHeader.Value = "close"
			}
		}
	}
	ctx.Response.AddHeader(connHeader)
}
