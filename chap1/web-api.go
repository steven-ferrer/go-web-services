package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

    "github.com/gorilla/mux"
)

type API struct {
	Message string `json:message`
}

func logger(r *http.Request){
    log.Println(r.Method, r.URL.Path)
}

func helloFunc(w http.ResponseWriter, r *http.Request){
    logger(r)
    urlParams := mux.Vars(r)
    name := urlParams["user"]

    helloMessage := "Hello, " + name

    message := API{Message:helloMessage}
    output, err := json.Marshal(message)

    if err != nil {
        fmt.Fprintf(w, "Something went wrong...")
    }

    fmt.Fprintf(w, string(output))
}

func main() {
    router := mux.NewRouter()
    router.HandleFunc("/api/{user:[0-9]+}", helloFunc)
    http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
