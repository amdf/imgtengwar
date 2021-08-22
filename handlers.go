package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
func NewTengwarImage() (ti *TengwarImage) {
	ti = new(TengwarImage)
	ti.Conv.InitDefault()
	ti.InitFonts()
	return
}

//InitFonts initalize fonts
func (ti *TengwarImage) InitFonts() {
	ti.fonts = make(map[string]*truetype.Font)

	for _, filename := range fontfiles {
		// Read the font data.
		fontBytes, err := ioutil.ReadFile(filename)
		if err == nil {
			f, err := freetype.ParseFont(fontBytes)
			if err == nil {
				ti.fonts[filename] = f
			}
		}
	}
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

	s, _ := ti.Conv.Convert(text)

	lines := strings.Split(s, "\n")

	ti.textToImage(lines, 72, w)
}
