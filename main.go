package main

import (
	"log"
	"os"

	"github.com/amdf/imgtengwar/internal/render"
	"github.com/gin-gonic/gin"
)

func main() {
	render.Init()

	server := TengwarImageServer{}
	server.Init()

	router := gin.Default()

	router.GET("/text", server.ConvertText)
	router.GET("/img", server.ConvertImage)

	port := os.Getenv("IMGTENGWAR_PORT")
	log.Print("Listen on port ", port)

	log.Fatal(router.Run("localhost:" + port))
}
