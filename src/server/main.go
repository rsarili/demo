package main

import (
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			log.Printf("%s: %s\n", k, v)
		}
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			log.Printf("%s: %s\n", k, v)
		}
		
		defer r.Body.Close()
		body_bytes, _ := io.ReadAll(r.Body)
		
		w.Write(body_bytes)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
