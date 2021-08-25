package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/amdf/imgtengwar/internal/render"
	"github.com/amdf/rustengwar"
	"github.com/gin-gonic/gin"
)

//TengwarImageServer contains handlers
type TengwarImageServer struct {
	TextConverter rustengwar.Converter
}

//Init server
func (server *TengwarImageServer) Init() {
	err := server.TextConverter.InitDefault()
	if err != nil {
		panic("converter is not initialized")
	}
}

//ConvertText returns converted text
func (server *TengwarImageServer) ConvertText(c *gin.Context) {
	text := c.Query("text")

	s, _ := server.TextConverter.Convert(text)

	c.String(http.StatusOK, "%s", s)
}

//ConvertImage shows image from converted text
func (server *TengwarImageServer) ConvertImage(c *gin.Context) {
	text := c.Query("text")
	size := c.Query("size")

	iSize, err := strconv.Atoi(size)

	if err != nil || iSize <= 0 {
		iSize = 36
	}

	s, _ := server.TextConverter.Convert(text)

	lines := strings.Split(s, "\n")

	err = render.ToPNG(lines, "tngan.ttf", float64(iSize), c.Writer)

	if err != nil {
		c.String(http.StatusNoContent, "error render.ToPNG")
	}
}
