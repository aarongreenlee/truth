// Package main is a basic example of how to perform integration testing in-process
// using the Truth package. This method allows test coverage to be calculated by Go.
package main

import (
	"fmt"
	"io"
	"net/http"
)

// To perform integration tests without actually going out to the network the server
// mux needs to be accessible to the Truth package. The tests are in this package and will
// have access to this un-exported mux.
var mux *http.ServeMux

func main() {
	fmt.Printf("Starting `github.com/aarongreenlee/truth/examples/basic`\n")

	// We'll separate configuration from mounting/listening so we can test
	// without actually going over the wire.
	configure()

	http.ListenAndServe(":65432", mux)
}

// Because we separated configuration from mounting/listening we can test
// the behavior of the server without going over the wire.
func configure() {
	mux = http.NewServeMux()

	mux.HandleFunc("/helloworld", func(res http.ResponseWriter, r *http.Request) {
		io.WriteString(res, "Hello world!")
	})
}
