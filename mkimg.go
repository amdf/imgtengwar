package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/amdf/rustengwar"
)

//TengwarImage handles img creation
type TengwarImage struct {
	Conv rustengwar.Converter
}

//NewTengwarImage creates TengwarImage
func NewTengwarImage() (ti *TengwarImage) {
	ti = new(TengwarImage)
	ti.Conv.InitDefault()
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

	//TODO: w.Header().Set("Content-Type", "")
	s, _ := ti.Conv.Convert(text)
	w.Write([]byte(s))
}
