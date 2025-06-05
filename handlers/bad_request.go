package handlers

import (
	"fmt"
	"gophttp/http"
)

func BadRequestHandler(ctx http.Context) error {
	ctx.Response.Status = http.StatusBadRequest
	//TODO: write reason for bad request
	var body string
	if reason, ok := ctx.AdditionalData["BadRequestReason"]; ok {
		body = fmt.Sprintf("Bad request: %v", reason)
	} else {
		body = "Bad request"
	}
	ctx.Response.Body = body
	ctx.Response.AddHeader(http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}
