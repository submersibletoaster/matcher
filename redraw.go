package main


import (
	"image"
	"image/color"
	"github.com/Nykakin/quantize"
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
	preview.Show(palette)
	return palette
}
