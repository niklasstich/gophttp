package server

import (
	"gophttp/handlers"
	"gophttp/http"
)

type RouteHandlerCollection interface {
	GetRoute(method http.Method) handlers.Handler
	DeleteRoute(method http.Method)
	InsertRoute(method http.Method, handler handlers.Handler)
}

type routeHandlers struct {
	handlers map[http.Method]handlers.Handler
}

func NewRouteHandlers() RouteHandlerCollection {
	return &routeHandlers{handlers: make(map[http.Method]handlers.Handler)}
}

func (r routeHandlers) GetRoute(method http.Method) handlers.Handler {
	return r.handlers[method]
}

func (r routeHandlers) DeleteRoute(method http.Method) {
	delete(r.handlers, method)
}

func (r routeHandlers) InsertRoute(method http.Method, handler handlers.Handler) {
	r.handlers[method] = handler
}
