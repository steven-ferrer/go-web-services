package main

import (
	"fmt"
	"log"
	"net/http"
)

func proxyMe(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func main() {
	http.HandleFunc("/hello", proxyMe)
	http.HandleFunc("/world", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("World!"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
