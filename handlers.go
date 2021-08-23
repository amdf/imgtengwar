package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/amdf/rustengwar"
	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
)

var (
	fontfiles = []string{"tngan.ttf", "tngani.ttf"}
)

//TengwarImage handles img creation
type TengwarImage struct {
	Conv  rustengwar.Converter
	fonts map[string]*truetype.Font
}

//NewTengwarImage creates TengwarImage
func NewTengwarImage() (ti *TengwarImage, err error) {
	ti = new(TengwarImage)
	err = ti.Conv.InitDefault()
	if nil == err {
		err = ti.InitFonts()
	}
	return
}

//InitFonts initalize fonts
func (ti *TengwarImage) InitFonts() (err error) {
	ti.fonts = make(map[string]*truetype.Font)

	for _, filename := range fontfiles {
		// Read the font data.
		fontBytes, errFile := ioutil.ReadFile(filename)
		if errFile == nil {
			f, errFont := freetype.ParseFont(fontBytes)
			if errFont == nil {
				ti.fonts[filename] = f
			}
		}
	}

	if 0 == len(ti.fonts) {
		err = errors.New("no fonts")
	}

	return
}

func (ti TengwarImage) getSingleParam(req *http.Request, key string) string {
	keys, ok := req.URL.Query()[key]

	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

//ConvertText shows converted text
func (ti *TengwarImage) ConvertText(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("expect method GET /text/, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	log.Printf("handling ConvertText at %s\n", req.URL.Path)

	text := ti.getSingleParam(req, "text")

	s, _ := ti.Conv.Convert(text)
	w.Write([]byte(s))
}

//ConvertImage shows image from converted text
func (ti *TengwarImage) ConvertImage(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, fmt.Sprintf("expect method GET /img/, got %v", req.Method), http.StatusMethodNotAllowed)
		return
	}

	log.Printf("handling ConvertImage at %s\n", req.URL.Path)

	text := ti.getSingleParam(req, "text")
	size := ti.getSingleParam(req, "size")

	iSize, err := strconv.Atoi(size)

	if err != nil || iSize <= 0 {
		iSize = 36
	}

	s, _ := ti.Conv.Convert(text)

	lines := strings.Split(s, "\n")

	err = ti.textToImage(lines, float64(iSize), w)

	if err != nil {
		http.Error(w, "ConvertImage error", http.StatusNoContent)
	}
}
