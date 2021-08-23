package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	ti, err := NewTengwarImage()
	if err != nil {
		log.Fatal("init failed")
	}

	mux.HandleFunc("/text/", ti.ConvertText)
	mux.HandleFunc("/img/", ti.ConvertImage)
	port := os.Getenv("IMGTENGWAR_PORT")
	log.Print("Listen on port ", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, mux))
}
