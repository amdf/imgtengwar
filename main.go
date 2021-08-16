package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	ti := NewTengwarImage()
	mux.HandleFunc("/text/", ti.ConvertText)
	port := os.Getenv("IMGTENGWAR_PORT")
	log.Print("Listen on port ", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, mux))
}
