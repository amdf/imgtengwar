package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	ti := TengwarImage{}
	ti.Init()
	mux.HandleFunc("/txt/", ti.ConvertText)
	port := os.Getenv("IMGTENGWAR_PORT")
	log.Print("Listen on port ", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, mux))
}
