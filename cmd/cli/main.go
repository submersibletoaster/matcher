package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/submersibletoaster/matcher"
)

var cellX = flag.Int("w", 8, "Cell width from source image. (A cell becomes 1 glyph)")
var cellY = flag.Int("h", 16, "Cell height from source image")

func main() {
	flag.Parse()
	srcFile := flag.Arg(0)
	srcIo, err := os.Open(srcFile)
	srcImg, _, err := image.Decode(srcIo)
	if err != nil {
		panic(err)
	}
	if srcImg.ColorModel() != color.RGBAModel {
		old := srcImg
		replace := image.NewRGBA(old.Bounds())
		draw.Draw(replace, old.Bounds(), old, image.ZP, draw.Src)
		srcImg = replace
	}

	pal := matcher.PickPalette(srcImg, 256)

	// TODO - be able to downsample rather than register the font glyph size directly as cell-size
	cells, expectCells := matcher.SliceImage(srcImg, image.Rect(0, 0, *cellX, *cellY), pal)

	diverse := make(map[string]uint)
	cb := func(c string, bg color.Color, fg color.Color) {
		diverse[c]++
	}
	//previewImage(image.NewRGBA(srcImg.Bounds()),cells,cb)
	matcher.WriteANSI(os.Stdout, cells, cb)

	fmt.Fprintf(os.Stderr, "\nDiversity = %d, %+v\n", len(diverse), diverse)
	fmt.Fprintf(os.Stderr, "Cells: %d\n", expectCells)
}
