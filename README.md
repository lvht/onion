# Onion

Onion is a for of https://github.com/urfave/negroni, which is is an idiomatic
approach to web middleware in Go. It is tiny, non-intrusive,
and encourages use of `net/http` Handlers.

Onion only use the http.Handler interface, and does not shift any middleware.

If you like the idea of [Martini](https://github.com/go-martini/martini), but
you think it contains too much magic, then Onion is a great fit.

## Getting Started

After installing Go and setting up your
[GOPATH](http://golang.org/doc/code.html#GOPATH), create your first `.go` file.
We'll call it `server.go`.

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "net/http"

  "github.com/lvht/onion"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := onion.New()
  n.UseHandler(mux)

  http.ListenAndServe(":3000", n)
}
```

Then install the Onion package:

```
go get github.com/lvht/onion
```

Then run your server:

```
go run server.go
```

You will now have a Go `net/http` webserver running on `localhost:3000`.

## Is Onion a Framework?

Onion is **not** a framework. It is a middleware-focused library that is
designed to work directly with `net/http`.

## Routing?

Onion is BYOR (Bring your own Router). The Go community already has a number
of great http routers available, and Onion tries to play well with all of them
by fully supporting `net/http`. For instance, integrating with [Gorilla Mux]
looks like so:

``` go
router := mux.NewRouter()
router.HandleFunc("/", HomeHandler)

n := onion.New(Middleware1, Middleware2)
// Or use a middleware with the Use() function
n.Use(Middleware3)
// router goes last
n.UseHandler(router)

http.ListenAndServe(":3001", n)
```

## Handlers

Onion provides a bidirectional middleware flow. This is done through the
`onion.Handler` interface:

``` go
type Handler interface {
  ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler)
}
```

If a middleware hasn't already written to the `ResponseWriter`, it should call
the next `http.Handler` in the chain to yield to the next middleware
handler.  This can be used for great good:

``` go
func MyMiddleware(rw http.ResponseWriter, r *http.Request, next http.Handler) {
  // do some stuff before
  next.ServeHTTP(rw, r)
  // do some stuff after
}
```

And you can map it to the handler chain with the `Use` function:

``` go
n := onion.New()
n.Use(onion.HandlerFunc(MyMiddleware))
```

You can also map plain old `http.Handler`s:

``` go
n := onion.New()

mux := http.NewServeMux()
// map your routes

n.UseHandler(mux)

http.ListenAndServe(":3000", n)
```

## `With()`

Onion has a convenience function called `With`. `With` takes one or more
`Handler` instances and returns a new `Onion` with the combination of the
receiver's handlers and the new handlers.

```go
// middleware we want to reuse
common := onion.New()
common.Use(MyMiddleware1)
common.Use(MyMiddleware2)

// `specific` is a new onion with the handlers from `common` combined with the
// the handlers passed in
specific := common.With(
	SpecificMiddleware1,
	SpecificMiddleware2
)
```

## Run

In general, you will want to use `net/http` methods and pass `onion` as a
`Handler`, as this is more flexible, e.g.:

<!-- { "interrupt": true } -->
``` go
package main

import (
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/lvht/onion"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Welcome to the home page!")
  })

  n := onion.New()
  n.UseHandler(mux)

  s := &http.Server{
    Addr:           ":8080",
    Handler:        n,
    ReadTimeout:    10 * time.Second,
    WriteTimeout:   10 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }
  log.Fatal(s.ListenAndServe())
}
```

## Route Specific Middleware

If you have a route group of routes that need specific middleware to be
executed, you can simply create a new Onion instance and use it as your route
handler.

``` go
router := mux.NewRouter()
adminRoutes := mux.NewRouter()
// add admin routes here

// Create a new onion for the admin middleware
router.PathPrefix("/admin").Handler(onion.New(
  Middleware1,
  Middleware2,
  onion.Wrap(adminRoutes),
))
```

If you are using [Gorilla Mux], here is an example using a subrouter:

``` go
router := mux.NewRouter()
subRouter := mux.NewRouter().PathPrefix("/subpath").Subrouter().StrictSlash(true)
subRouter.HandleFunc("/", someSubpathHandler) // "/subpath/"
subRouter.HandleFunc("/:id", someSubpathHandler) // "/subpath/:id"

// "/subpath" is necessary to ensure the subRouter and main router linkup
router.PathPrefix("/subpath").Handler(onion.New(
  Middleware1,
  Middleware2,
  onion.Wrap(subRouter),
))
```

`With()` can be used to eliminate redundancy for middlewares shared across
routes.

``` go
router := mux.NewRouter()
apiRoutes := mux.NewRouter()
// add api routes here
webRoutes := mux.NewRouter()
// add web routes here

// create common middleware to be shared across routes
common := onion.New(
	Middleware1,
	Middleware2,
)

// create a new onion for the api middleware
// using the common middleware as a base
router.PathPrefix("/api").Handler(common.With(
  APIMiddleware1,
  onion.Wrap(apiRoutes),
))
// create a new onion for the web middleware
// using the common middleware as a base
router.PathPrefix("/web").Handler(common.With(
  WebMiddleware1,
  onion.Wrap(webRoutes),
))
```

## Essential Reading for Beginners of Go & Onion

* [Using a Context to pass information from middleware to end handler](http://elithrar.github.io/article/map-string-interface/)
* [Understanding middleware](https://mattstauffer.co/blog/laravel-5.0-middleware-filter-style)

## About

Onion is obsessively designed by none other than the [Code Gangsta](https://codegangsta.io/)

[Gorilla Mux]: https://github.com/gorilla/mux
