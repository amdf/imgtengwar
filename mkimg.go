package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/amdf/rustengwar"
	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font"
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
	fontfiles := []string{"tngan.ttf", "tngani.ttf"}

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

func (ti TengwarImage) textToImage(text string, size float64, b io.Writer) (err error) {
	dpi := float64(72)
	fontfile := "tngani.ttf"
	hinting := "none"
	spacing := float64(1.5)
	wonb := false

	f := ti.fonts[fontfile]

	// Initialize the context.
	fg, bg := image.Black, image.White
	ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	if wonb {
		fg, bg = image.White, image.Black
		ruler = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	switch hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	// Draw the guidelines.
	for i := 0; i < 200; i++ {
		rgba.Set(10, 10+i, ruler)
		rgba.Set(10+i, 10, ruler)
	}

	// Draw the text.
	pt := freetype.Pt(10, 10+int(c.PointToFixed(size)>>6))

	_, err = c.DrawString(text, pt)
	if err != nil {
		//log.Println(err)
		return
	}
	pt.Y += c.PointToFixed(size * spacing)

	err = png.Encode(b, rgba)
	if err != nil {
		//log.Println(err)
		//os.Exit(1)
	}

	return
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
	ti.textToImage(s, 72, w)
}
