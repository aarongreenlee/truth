package main

import (
	"net/http"
	"net/url"

	"github.com/aarongreenlee/truth"
	"github.com/dimfeld/httptreemux"
)

type (
	MuxHandler func(http.ResponseWriter, *http.Request, url.Values)

	// Multiplexer defines an interface for the core system multiplexer.
	Multiplexer interface {
		http.Handler

		// Handle sets the MuxHandler for a given HTTP method and path using the
		// Truth definition and the provided handler.
		Handle(def truth.Definition, handle MuxHandler)

		HandleNotFound(handle MuxHandler)
	}

	mux struct {
		router *httptreemux.TreeMux
	}
)

func NewMux() Multiplexer {
	return &mux{
		router: httptreemux.New(),
	}
}

// ServeHTTP is the function called back by the underlying HTTP server to handle incoming requests.
func (m *mux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.router.ServeHTTP(rw, req)
}

// Handle parses the truth definition for the handler and registers the route with the multiplexer.
func (m *mux) Handle(def truth.Definition, h MuxHandler) {
	handle := func(rw http.ResponseWriter, req *http.Request, htparams map[string]string) {
		params := req.URL.Query()
		for n, p := range htparams {
			params.Set(n, p)
		}

		h(rw, req, params)
	}

	m.router.Handle(def.Method, def.Path, handle)
}

// HandleNotFound sets the MuxHandler invoked for requests that don't match any
// handler registered with Handle.
func (m *mux) HandleNotFound(h MuxHandler) {
	m.router.NotFoundHandler = func(rw http.ResponseWriter, req *http.Request) {
		h(rw, req, nil)
	}

	m.router.MethodNotAllowedHandler = func(rw http.ResponseWriter, req *http.Request, methods map[string]httptreemux.HandlerFunc) {
		h(rw, req, nil)
	}
}
