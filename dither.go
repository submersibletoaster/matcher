package matcher

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/segment"
	"github.com/esimov/colorquant"
)

var ThresholdPalette = color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}}

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
	return dither["Sierra-3"].Quantize(src, dst, n, true, false)
}

func QuantizeToPalette(src image.Image, pal color.Palette, n int) image.Image {
	b := src.Bounds()
	dst := image.NewPaletted(image.Rect(0, 0, b.Dx(), b.Dy()), pal)
	//dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	return colorquant.NoDither.Quantize(src, dst, n, false, false)
}

func DynamicThreshold(src image.Image) image.Image {
	gray := effect.Grayscale(src)
	darkest := uint8(0xff)
	lightest := uint8(0x00)
	hist := make([]uint32, 0xff+1)

	highest := uint8(0)
	second := uint8(0)
	// cheat and skip alpha
	for pos := 0; pos < len(gray.Pix); pos += 4 {
		//	for _, p := range gray.Pix {
		p := gray.Pix[pos]
		if p > lightest {
			lightest = p
		}
		if p < darkest {
			darkest = p
		}
		hist[p]++
		if hist[p] > hist[highest] {
			highest = p
		}
		if hist[p] > hist[second] && hist[p] < hist[highest] {
			second = p
		}

	}

	midpoint := darkest + (lightest-darkest)/2
	//guess := (midpoint + highest + 64) / 3
	guess := (highest + second) / 2
	/*
		fmt.Fprintf(os.Stderr, "mp: %d highest density %d\n", midpoint, highest)
		fmt.Fprintf(os.Stderr, "%+v\n", hist)
		fmt.Fprintf(os.Stderr, "%+v\n", gray.Pix)
	*/
	fmt.Fprintf(os.Stderr, "Light:% 3d, Dark:% 3d, Most:% 3d 2ndMost:% 3d MidPoint:% 3d Guessed : %d\n",
		lightest, darkest, highest, second, guess, midpoint)
	//bw := segment.Threshold(gray, guess)
	bw := segment.Threshold(gray, guess)
	//bwIdx := image.NewPaletted(bw.Bounds(), ThresholdPalette)
	//draw.Draw(bwIdx, bw.Bounds(), bw, bw.Bounds().Min, draw.Src)
	return bw
	//bwIndexed := matcher.QuantizeToPalette(bw, matcher.ThresholdPalette, 2)
	//return bwIndexed
	//	return gray
}
