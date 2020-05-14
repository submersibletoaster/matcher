package matcher

import (
	"image"
	"image/color"

	"github.com/esimov/colorquant"
)

var dither map[string]colorquant.Dither = map[string]colorquant.Dither{
	"FloydSteinberg": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0},
			[]float32{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0},
			[]float32{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0},
		},
	},
	"Burkes": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 8.0 / 32.0, 4.0 / 32.0},
			[]float32{2.0 / 32.0, 4.0 / 32.0, 8.0 / 32.0, 4.0 / 32.0, 2.0 / 32.0},
			[]float32{0.0, 0.0, 0.0, 0.0, 0.0},
			[]float32{4.0 / 32.0, 8.0 / 32.0, 0.0, 0.0, 0.0},
		},
	},
	"Stucki": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 8.0 / 42.0, 4.0 / 42.0},
			[]float32{2.0 / 42.0, 4.0 / 42.0, 8.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0},
			[]float32{1.0 / 42.0, 2.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0, 1.0 / 42.0},
		},
	},
	"Atkinson": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 1.0 / 8.0, 1.0 / 8.0},
			[]float32{1.0 / 8.0, 1.0 / 8.0, 1.0 / 8.0, 0.0},
			[]float32{0.0, 1.0 / 8.0, 0.0, 0.0},
		},
	},
	"Sierra-3": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 5.0 / 32.0, 3.0 / 32.0},
			[]float32{2.0 / 32.0, 4.0 / 32.0, 5.0 / 32.0, 4.0 / 32.0, 2.0 / 32.0},
			[]float32{0.0, 2.0 / 32.0, 3.0 / 32.0, 2.0 / 32.0, 0.0},
		},
	},
	"Sierra-2": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 0.0, 4.0 / 16.0, 3.0 / 16.0},
			[]float32{1.0 / 16.0, 2.0 / 16.0, 3.0 / 16.0, 2.0 / 16.0, 1.0 / 16.0},
			[]float32{0.0, 0.0, 0.0, 0.0, 0.0},
		},
	},
	"Sierra-Lite": colorquant.Dither{
		[][]float32{
			[]float32{0.0, 0.0, 2.0 / 4.0},
			[]float32{1.0 / 4.0, 1.0 / 4.0, 0.0},
			[]float32{0.0, 0.0, 0.0},
		},
	},
}

func DitherToPalette(src image.Image, pal color.Palette, n int) image.Image {
	b := src.Bounds()
	dst := image.NewPaletted(image.Rect(0, 0, b.Dx(), b.Dy()), pal)
	return dither["Stucki"].Quantize(src, dst, n, true, true)
}
