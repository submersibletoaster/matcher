package main


import (
	"fmt"
	"image"
	"image/color"
	"io"
	//"image/draw"
	"github.com/Nykakin/quantize"
	"github.com/rivo/duplo"
	"sort"
	"github.com/joshdk/preview"
)

// Pick the overall image palette
func pickPalette(img image.Image,num int) (color.Palette) {
	q := quantize.NewHierarhicalQuantizer()
	colors,err := q.Quantize(img,num)
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

func findBestStructure(cell image.Image,store *duplo.Store) (*duplo.Match,color.Color,color.Color) {
	// github.com/disintegration/imaging
	//inverted := draw.Invert(cell)
	monoPal := pickPalette(cell,2)
	//monoImg := image.NewPaletted(cell.Bounds(),monoPal)
	//draw.Draw(monoImg,cell.Bounds(),cell,image.ZP,draw.Src)
	monoImg := ditherToPalette(cell,monoPal,2)

	// *actual* representative colours
	bgCol := monoPal[0]
	fgCol := monoPal[1]
	// clobber with black/white to match the glyph database
	monoPal[0] = color.RGBA{0,0,0,255}
	monoPal[1] = color.RGBA{255,255,255,255}
	if ! shown {
		fmt.Printf("MonoPal: %v\n", monoPal)
		fmt.Printf("MonoImg: %v\n", monoImg)
		shown = true
	}


	hash,_ := duplo.CreateHash(monoImg)

	//hash,_ := duplo.CreateHash(cell)
	matches := store.Query(hash)
	sort.Sort(matches)
	//d,n := fgDensity(monoImg.(*image.Paletted))
	//fmt.Printf("density: %d ; %s ; %2.2f\n", d, string(rune(matches[0].ID.(int32))), n )
	return matches[0],bgCol,fgCol
}

func fgDensity(src *image.Paletted) (count uint,norm float64) {
	limit := src.Bounds().Dx() * src.Bounds().Dy()
	count = 0
	for _,v := range src.Pix {
		count += uint(v)
	}
	return count, float64(count)/float64(limit)
}


func writeANSI(w io.Writer,cells chan Cell) () {

}

