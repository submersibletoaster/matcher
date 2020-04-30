package main


import (
	"image"
	"image/color"
	"github.com/Nykakin/quantize"
	"github.com/rivo/duplo"
	"sort"
	"github.com/joshdk/preview"
)

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


func findBestStructure(cell image.Image,store *duplo.Store) (*duplo.Match) {
	// github.com/disintegration/imaging
	//inverted := draw.Invert(cell)
	hash,_ := duplo.CreateHash(cell)
	matches := store.Query(hash)
	sort.Sort(matches)
	return matches[0]


}
