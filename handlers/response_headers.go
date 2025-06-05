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

var ServerHeader, ConnectionHeader http.Header

func init() {
	ServerHeader = http.Header{
		Name:  "Server",
		Value: fmt.Sprintf("%s/%s", SERVER_NAME, VERSION),
	}
	ConnectionHeader = http.Header{
		Name:  "Connection",
		Value: "close",
	}
}

func ResponseHeadersHandler(ctx http.Context) error {
	//Server
	ctx.Response.AddHeader(ServerHeader)
	//Date
	ctx.Response.AddHeader(http.Header{
		Name:  "Date",
		Value: common.ToHttpDateFormat(time.Now()),
	})
	ctx.Response.AddHeader(ConnectionHeader)
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
	return nil
}
