package main

import (
	"log"
	"main/channels"
	"main/mutexes"
	"net/http"
)

func main() {
	http.HandleFunc("/channels", channels.WsHandler)
	http.HandleFunc("/mutexes", mutexes.WsHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Println("Server started on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
