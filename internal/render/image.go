package render

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	padding = 10
	spacing = float64(1.5)
)

var (
	fonts     map[string]*truetype.Font
	fontfiles = []string{"tngan.ttf", "tngani.ttf"}
)

//Init initalize fonts for rendering
func Init() (err error) {
	fonts = make(map[string]*truetype.Font)

	for _, filename := range fontfiles {
		// Read the font data.
		fontBytes, errFile := ioutil.ReadFile(filename)
		if errFile == nil {
			f, errFont := freetype.ParseFont(fontBytes)
			if errFont == nil {
				fonts[filename] = f
			}
		}
	}

	if 0 == len(fonts) {
		panic("no fonts")
	}

	return
}

//ToPNG converts lines of text to PNG image
func ToPNG(text []string, fontfile string, size float64, w io.Writer) (err error) {
	dpi := float64(72)

	hinting := "none"

	f, ok := fonts[fontfile]
	if !ok {
		err = errors.New("unknown font")
		return
	}

	// Initialize the context.
	fg, bg := image.White, image.Black

	rgba := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

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

	startY := padding + int(c.PointToFixed(size)>>6)

	ptLeft := freetype.Pt(10, startY)
	fmt.Println("start from: ", ptLeft.X, " ", ptLeft.Y)

	var ptRight fixed.Point26_6
	var maxRightX int

	fmt.Println("drawPoint 1 Y: ", ptRight.Y)

	//Draw lines
	for _, s := range text {
		ptRight, err = c.DrawString(s, ptLeft)
		fmt.Println("string: ", ptRight.X, " ", ptRight.Y)

		if maxRightX < ptRight.X.Ceil() {
			maxRightX = ptRight.X.Ceil()
		}

		if err != nil {
			return
		}
		ptLeft.Y += c.PointToFixed(size * spacing)
	}

	if maxRightX > padding {
		// Crop image
		err = png.Encode(w, rgba.SubImage(
			image.Rectangle{
				Min: image.Point{X: 0, Y: 0},
				Max: image.Point{
					X: padding + maxRightX,
					Y: ptRight.Y.Ceil() + int(size),
				},
			},
		))
	} else {
		err = errors.New("resulting image too small")
	}

	return
}
