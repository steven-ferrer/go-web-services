package main

import (
	"fmt"
	"net/http"
	"strconv"

	"golang.org/x/net/websocket"
)

var address = ":12345"

func EchoLengthServer(ws *websocket.Conn) {
	var msg string

	for {
		websocket.Message.Receive(ws, &msg)
		fmt.Println("Got message: ", msg)
		length := len(msg)

		if err := websocket.Message.Send(ws, strconv.FormatInt(int64(length), 10)); err != nil {
			fmt.Println("Can't send message length")
			break
		}
	}
}

func websocketListen() {
	http.Handle("/length", websocket.Handler(EchoLengthServer))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func main() {
	//resource files
	http.Handle("/", http.FileServer(http.Dir("../../jq")))
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket.html")
	})

	websocketListen()
}
