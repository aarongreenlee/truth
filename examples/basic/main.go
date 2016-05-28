package main

import (
	"io"
	"net/http"
	"fmt"
)

func textReply(res http.ResponseWriter, r *http.Request) {
	fmt.Printf("Serving the `textReply` endpoint\n")

	res.WriteHeader(200)
	io.WriteString(res, "Hello world!")
}

var mux *http.ServeMux

func main() {
	fmt.Printf("Starting `github.com/aarongreenlee/truth/examples/basic`\n")
	bootstrapApp()
	mount()
}

func bootstrapApp() {
	mux = http.NewServeMux()
	mux.HandleFunc("/helloworld", textReply)

}

func mount() {
	http.ListenAndServe(":8000", mux)
}