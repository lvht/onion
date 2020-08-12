// Package onion is a for of https://github.com/urfave/negroni, which is
// an idiomatic approach to web middleware in Go. It is tiny, non-intrusive,
// and encourages use of net/http Handlers.
//
// Onion only use the http.Handler interface, and does not shift any middleware.
//
// For a full guide visit http://github.com/lvht/onion
//
//  package main
//
//  import (
//    "net/http"
//    "fmt"
//
//    "github.com/lvht/onion"
//  )
//
//  func main() {
//    mux := http.NewServeMux()
//    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
//      fmt.Fprintf(w, "Welcome to the home page!")
//    })
//
//    n := onion.New()
//    n.UseHandler(mux)
//  }
package onion
