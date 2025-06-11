package handlers

import (
	"gophttp/http"
)

type IdentityHandler struct {
}

func (i IdentityHandler) HandleRequest(ctx http.Context) error {
	ctx.Response.AddHeader(http.Header{
		Name:  "Content-Encoding",
		Value: "identity",
	})
	return nil
}
