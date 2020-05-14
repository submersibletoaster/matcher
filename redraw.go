package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"

	//"image/draw"
	"sort"

	"github.com/Nykakin/quantize"
	"github.com/joshdk/preview"
	"github.com/rivo/duplo"
)

// Pick the overall image palette
func pickPalette(img image.Image, num int) color.Palette {
	q := quantize.NewHierarhicalQuantizer()
	colors, err := q.Quantize(img, num)
	if err != nil {
		panic(err)
	}

	palette := make([]color.Color, len(colors))
	for index, clr := range colors {
		palette[index] = clr
	}

	// Display our new palette
	if false {
		preview.Show(palette)
	}
	return palette
}

var shown bool = false

func findBestStructure(cell image.Image, store *duplo.Store) (*duplo.Match, color.Color, color.Color) {
	// github.com/disintegration/imaging
	//inverted := draw.Invert(cell)
	monoPal := pickPalette(cell, 2)
	//monoImg := image.NewPaletted(cell.Bounds(),monoPal)
	//draw.Draw(monoImg,cell.Bounds(),cell,image.ZP,draw.Src)
	monoImg := ditherToPalette(cell, monoPal, 2)

	// *actual* representative colours
	bgCol := monoPal[0]
	fgCol := monoPal[1]
	// clobber with black/white to match the glyph database
	monoPal[0] = color.RGBA{0, 0, 0, 255}
	monoPal[1] = color.RGBA{255, 255, 255, 255}
	if !shown {
		fmt.Printf("MonoPal: %v\n", monoPal)
		fmt.Printf("MonoImg: %v\n", monoImg)
		shown = true
	}

	hash, _ := duplo.CreateHash(monoImg)

	monoPal[1] = color.RGBA{0, 0, 0, 255}
	monoPal[0] = color.RGBA{255, 255, 255, 255}
	invhash, _ := duplo.CreateHash(monoImg)

	//hash,_ := duplo.CreateHash(cell)
	matches := store.Query(hash)
	sort.Sort(matches)
	invmatches := store.Query(invhash)
	sort.Sort(invmatches)
	if invmatches[0].Score < matches[0].Score {
		return invmatches[0], fgCol, bgCol
	}
	//d,n := fgDensity(monoImg.(*image.Paletted))
	//fmt.Printf("density: %d ; %s ; %2.2f\n", d, string(rune(matches[0].ID.(int32))), n )
	return matches[0], bgCol, fgCol
}

func fgDensity(src *image.Paletted) (count uint, norm float64) {
	limit := src.Bounds().Dx() * src.Bounds().Dy()
	count = 0
	for _, v := range src.Pix {
		count += uint(v)
	}
	return count, float64(count) / float64(limit)
}

func writeANSI(w io.Writer, cells chan Cell) {

}

func previewImage(output draw.Image, cells chan Cell,cb func(string,color.Color,color.Color)) {
	draw.Draw(output, output.Bounds(), image.Black, image.ZP, draw.Src)
	for cell := range cells {
		m, bg, fg := findBestStructure(cell.Image, fontStore)
		char := string(rune(m.ID.(int32)))
		draw.Draw(output, cell.Bounds, image.NewUniform(bg), image.ZP, draw.Src)
		myFont.DrawRune(output, cell.Bounds.Min.X, cell.Bounds.Min.Y, rune(m.ID.(int32)), fg)
		cb(char,bg,fg)
	}
	preview.Image(output)
}
