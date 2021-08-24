package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/amdf/rustengwar"
	"github.com/gin-gonic/gin"
	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
)

var (
	fontfiles = []string{"tngan.ttf", "tngani.ttf"}
)

//TengwarImageServer handles img creation
type TengwarImageServer struct {
	Conv  rustengwar.Converter
	fonts map[string]*truetype.Font
}

//NewTengwarImageServer creates TengwarImage
func NewTengwarImageServer() (ti *TengwarImageServer, err error) {
	ti = new(TengwarImageServer)
	err = ti.Conv.InitDefault()
	if nil == err {
		err = ti.InitFonts()
	}
	return
}

//InitFonts initalize fonts
func (server *TengwarImageServer) InitFonts() (err error) {
	server.fonts = make(map[string]*truetype.Font)

	for _, filename := range fontfiles {
		// Read the font data.
		fontBytes, errFile := ioutil.ReadFile(filename)
		if errFile == nil {
			f, errFont := freetype.ParseFont(fontBytes)
			if errFont == nil {
				server.fonts[filename] = f
			}
		}
	}

	if 0 == len(server.fonts) {
		err = errors.New("no fonts")
	}

	return
}

//ConvertText returns converted text
func (server *TengwarImageServer) ConvertText(c *gin.Context) {
	text := c.Query("text")

	s, _ := server.Conv.Convert(text)

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

	s, _ := server.Conv.Convert(text)

	lines := strings.Split(s, "\n")

	err = server.textToImage(lines, "tngan.ttf", float64(iSize), c.Writer)

	if err != nil {
		c.String(http.StatusNoContent, "ConvertImage error")
	}
}
