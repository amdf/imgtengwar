package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	server, err := NewTengwarImageServer()
	if err != nil {
		log.Fatal("init failed")
	}

	router := gin.Default()

	router.GET("/text", server.ConvertText)
	router.GET("/img", server.ConvertImage)

	port := os.Getenv("IMGTENGWAR_PORT")
	log.Print("Listen on port ", port)

	log.Fatal(router.Run("localhost:" + port))
}
