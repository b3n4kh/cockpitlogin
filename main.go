package main

import (
	"log"
	"net"
	"net/http"
)

func main() {
	file := "/run/cockpitlogin/socket"
	listener, err := net.Listen("unix", file)
	if err != nil {
		log.Fatalf("Could not listen on %s: %v", file, err)
		return
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello World"))
	})
	defer listener.Close()
	if err = http.Serve(listener, nil); err != nil {
		log.Fatalf("Could not start HTTP server: %v", err)
	}
	log.Printf("Started Listening on: %s", file)
}
