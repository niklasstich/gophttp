package handlers

import "gophttp/http"

type Handler interface {
	HandleRequest(ctx http.Context) error
}

type HandlerFunc func(ctx http.Context) error

func (h HandlerFunc) HandleRequest(ctx http.Context) error {
	return h(ctx)
}

type composedHandler struct {
	h1, h2 Handler
}

func (c composedHandler) HandleRequest(ctx http.Context) error {
	err := c.h1.HandleRequest(ctx)
	if err != nil {
		return err
	}
	err = c.h2.HandleRequest(ctx)
	return err
}

func ComposeHandlers(h1, h2 Handler) Handler {
	return composedHandler{h1, h2}
}

func NotFoundHandler(ctx http.Context) error {
	ctx.Response.Status = http.StatusNotFound
	ctx.Response.Body = "Page doesn't exist"
	ctx.Response.AddHeader(http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}

func InternalServerErrorHandler(ctx http.Context) error {
	ctx.Response.Status = http.StatusInternalServerError
	ctx.Response.Body = "Internal server error"
	ctx.Response.AddHeader(http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}
