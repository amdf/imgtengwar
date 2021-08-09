package main

import (
	"log"
	"net/http"

	"github.com/amdf/rustengwar"
)

//TengwarImage handles img creation
type TengwarImage struct {
	conv rustengwar.Converter
}

func (ti *TengwarImage) Init() {
	ti.conv.InitDefault()
}
func (ti *TengwarImage) ConvertText(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling GetImage at %s\n", req.URL.Path)
	//TODO: w.Header().Set("Content-Type", "")
	s, _ := ti.conv.Convert("прекрасно")
	w.Write([]byte(s))
}
