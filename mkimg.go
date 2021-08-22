package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"

	"github.com/goki/freetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func (ti TengwarImage) textToImage(text []string, size float64, b io.Writer) (err error) {
	dpi := float64(72)
	fontfile := "tngan.ttf"
	hinting := "none"
	spacing := float64(1.5)

	f := ti.fonts[fontfile]

	// Initialize the context.
	fg, bg := image.Black, image.White

	rgba := image.NewRGBA(image.Rect(0, 0, 1024, 768))
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

	startY := 10 + int(c.PointToFixed(size)>>6)

	ptLeft := freetype.Pt(10, startY)
	fmt.Println("start from: ", ptLeft.X, " ", ptLeft.Y)

	var ptRight fixed.Point26_6
	var maxRightX int

	fmt.Println("drawPoint 1 Y: ", ptRight.Y)
	for _, s := range text {
		ptRight, err = c.DrawString(s, ptLeft)
		fmt.Println("string: ", ptRight.X, " ", ptRight.Y)

		if maxRightX < ptRight.X.Ceil() {
			maxRightX = ptRight.X.Ceil()
		}

		if err != nil {
			log.Println(err)
			return
		}
		ptLeft.Y += c.PointToFixed(size * spacing)
	}

	// Crop image
	err = png.Encode(b, rgba.SubImage(
		image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: 10 + maxRightX, Y: ptRight.Y.Ceil() + int(size)},
		},
	))

	if err != nil {
		log.Println(err)
	}

	return
}
