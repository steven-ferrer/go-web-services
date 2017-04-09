package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

const (
	serverName   = "localhost"
	SSLPort      = ":10443"
	HTTPPort     = ":8080"
	SSLProtocol  = "https://"
	HTTPProtocol = "http://"
)

func secureRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HTTPS Requested")
	fmt.Fprintln(w, "You have arrived at port 443, but you are not yet secure.")
}

func redirectNonSecure(w http.ResponseWriter, r *http.Request) {
	log.Println("Non-secure request initiated, redirecting...")
	redirectURL := SSLProtocol + serverName + SSLPort + r.RequestURI
	fmt.Println(redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func main() {
	wg := sync.WaitGroup{}
	log.Println("Starting redirection server, try to access @http:")

	//increment WaitGroup counter
	//this blocks until the next goroutine calls wg.Done()
	wg.Add(1)
	go func() {
		http.ListenAndServe(HTTPPort, http.HandlerFunc(redirectNonSecure))
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		http.ListenAndServeTLS(SSLPort, "cert.pem", "key.pem",
			http.HandlerFunc(secureRequest))
		wg.Done()
	}()

	//wait for all the jobs to complete
	wg.Wait()
	/*http.HandleFunc("/", handler)
	log.Printf("About to listen on 10443. Go to https://127.0.0.1:10443/")
	err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	log.Fatal(err)*/
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}
