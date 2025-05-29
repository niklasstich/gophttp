package handlers

import "http/http"

type Handler interface {
	HandleRequest(req http.Request, resp *http.Response) error
}

type HandlerFunc func(req http.Request, resp *http.Response)

func (h HandlerFunc) HandleRequest(req http.Request, resp *http.Response) error {
	return h.HandleRequest(req, resp)
}

type composedHandler struct {
	h1, h2 Handler
}

func (c composedHandler) HandleRequest(req http.Request, resp *http.Response) error {
	err := c.h1.HandleRequest(req, resp)
	if err != nil {
		return err
	}
	err = c.h2.HandleRequest(req, resp)
	return err
}

func ComposeHandlers(h1, h2 Handler) Handler {
	return composedHandler{h1, h2}
}

func NotFoundHandler(_ http.Request, resp *http.Response) error {
	resp.Status = http.StatusNotFound
	resp.Body = "Page doesn't exist"
	resp.Headers = append(resp.Headers, http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}

func BadRequestHandler(req http.Request, resp *http.Response) error {
	resp.Status = http.StatusBadRequest
	//TODO: write reason for bad request
	resp.Body = "Bad request"
	resp.Headers = append(resp.Headers, http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}

func InternalServerErrorHandler(req http.Request, resp *http.Response) error {
	resp.Status = http.StatusInternalServerError
	//TODO: (maybe?) write reason for internal server error
	resp.Body = "Internal server error"
	resp.Headers = append(resp.Headers, http.Header{
		Name:  "MIME",
		Value: "text/plain",
	})
	return nil
}
