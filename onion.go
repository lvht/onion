package onion

import (
	"net/http"
)

// Handler handler is an interface that objects can implement to be registered to serve as middleware
// in the Onion middleware stack.
// ServeHTTP should yield to the next middleware in the chain by invoking the next http.HandlerFunc
// passed in.
//
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should not be invoked.
type Handler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Onion handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.Handler)

func (h HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	h(rw, r, next)
}

type middleware struct {
	handler Handler

	// next stores the next.ServeHTTP to reduce memory allocate
	next http.Handler
}

func (m middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if m.handler == nil {
		return
	}

	m.handler.ServeHTTP(rw, r, m.next)
}

// Wrap converts a http.Handler into a onion.Handler so it can be used as a Onion
// middleware. The next http.HandlerFunc is automatically called after the Handler
// is executed.
func Wrap(handler http.Handler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		handler.ServeHTTP(rw, r)
		next.ServeHTTP(rw, r)
	})
}

// WrapFunc converts a http.HandlerFunc into a onion.Handler so it can be used as a Onion
// middleware. The next http.HandlerFunc is automatically called after the Handler
// is executed.
func WrapFunc(handlerFunc http.HandlerFunc) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
		handlerFunc(rw, r)
		next.ServeHTTP(rw, r)
	})
}

// Onion is a stack of Middleware Handlers that can be invoked as an http.Handler.
// Onion middleware is evaluated in the order that they are added to the stack using
// the Use and UseHandler methods.
type Onion struct {
	middleware middleware
	handlers   []Handler
}

// New returns a new Onion instance with no middleware preconfigured.
func New(handlers ...Handler) *Onion {
	return &Onion{
		handlers:   handlers,
		middleware: build(handlers),
	}
}

// With returns a new Onion instance that is a combination of the onion
// receiver's handlers and the provided handlers.
func (n *Onion) With(handlers ...Handler) *Onion {
	currentHandlers := make([]Handler, len(n.handlers))
	copy(currentHandlers, n.handlers)
	return New(
		append(currentHandlers, handlers...)...,
	)
}

func (n *Onion) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	n.middleware.ServeHTTP(rw, r)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a Onion.
func (n *Onion) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	n.handlers = append(n.handlers, handler)
	n.middleware = build(n.handlers)
}

// UseFunc adds a Onion-style handler function onto the middleware stack.
func (n *Onion) UseFunc(handlerFunc func(rw http.ResponseWriter, r *http.Request, next http.Handler)) {
	n.Use(HandlerFunc(handlerFunc))
}

// UseHandler adds a http.Handler onto the middleware stack. Handlers are invoked in the order they are added to a Onion.
func (n *Onion) UseHandler(handler http.Handler) {
	n.Use(Wrap(handler))
}

// UseHandlerFunc adds a http.HandlerFunc-style handler function onto the middleware stack.
func (n *Onion) UseHandlerFunc(handlerFunc func(rw http.ResponseWriter, r *http.Request)) {
	n.UseHandler(http.HandlerFunc(handlerFunc))
}

// Returns a list of all the handlers in the current Onion middleware chain.
func (n *Onion) Handlers() []Handler {
	return n.handlers
}

func build(handlers []Handler) middleware {
	if len(handlers) == 0 {
		return middleware{}
	}

	var next middleware
	if len(handlers) > 1 {
		next = build(handlers[1:])
	}

	return middleware{handler: handlers[0], next: &next}
}
