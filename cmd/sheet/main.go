package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/segment"

	"github.com/submersibletoaster/matcher"
)

func main() {
	font := matcher.GetFont()
	lookup := matcher.GetLookup()
	density := matcher.GetDensity()
	chars := len(lookup)
	fmt.Printf("%d chars\n", chars)
	cols := int(math.Ceil(math.Sqrt(float64(chars))))
	sX := 8 // magic number - should inspect from font
	sY := font.GetHeight()
	r := image.Rect(0, 0, sX*cols, sY*cols)
	i := image.NewRGBA(r)

	allChars := make([]string, chars)
	n := 0
	for k, _ := range lookup {
		allChars[n] = k
		n++
	}

	n = 0
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	draw.Draw(i, i.Bounds(), image.NewUniform(black), image.ZP, draw.Src)

	for y := 0; y < r.Max.Y; y += sY {
		for x := 0; x < r.Max.X; x += sX {
			font.DrawString(i, x, y, allChars[n], white)
			n++
			if n >= chars {
				break
			}
		}
	}

	w, _ := os.Create("out.png")
	png.Encode(w, i)
	w.Close()

	flag.Parse()
	srcFile := flag.Arg(0)
	srcIo, _ := os.Open(srcFile)
	srcImg, _, _ := image.Decode(srcIo)
	if srcImg.ColorModel() != color.RGBAModel {
		old := srcImg
		replace := image.NewRGBA(old.Bounds())
		draw.Draw(replace, old.Bounds(), old, image.ZP, draw.Src)
		srcImg = replace
	}

	output := image.NewRGBA(srcImg.Bounds())
	thrOut := image.NewRGBA(srcImg.Bounds())

	pal := matcher.PickPalette(srcImg, 64)
	//hardpal := matcher.ThresholdPalette

	// TODO - be able to downsample rather than register the font glyph size directly as cell-size
	cells, _ := matcher.SliceImage(srcImg, image.Rect(0, 0, 8, 16), pal)
	draw.Draw(output, output.Bounds(), image.Black, image.ZP, draw.Src)
	for cell := range cells {
		/*		quant := matcher.QuantizeToPalette(cell.Image,pal,2)
				fmt.Fprintf(os.Stderr,"%+v\n", quant.(*image.Paletted).Palette )
				fmt.Fprintf(os.Stderr,"%+v\n", quant.(*image.Paletted).Pix )
				fmt.Fprintf(os.Stderr, "%#v\t\n", quant.Bounds())
				draw.Draw(output, cell.Bounds, quant, image.ZP  , draw.Src)
		*/
		seg := DynamicThreshold(cell.Image)
		//segment.Threshold(cell.Image, 64)
		draw.Draw(thrOut, seg.Bounds(), seg, seg.Bounds().Min, draw.Src)

		c, _, fg := matcher.FindBestMatch(seg, density)
		font.DrawString(output, seg.Bounds().Min.X, seg.Bounds().Min.Y, c, fg)

		//draw.Draw(output,cell.Bounds,cell.Image,cell.Bounds.Min, draw.Src)
	}

	ow, _ := os.Create("preview.png")
	png.Encode(ow, output)
	ow.Close()

	ot, _ := os.Create("threshold.png")
	png.Encode(ot, thrOut)
	ot.Close()

}

func DynamicThreshold(src image.Image) image.Image {
	gray := effect.Grayscale(src)
	darkest := uint8(0xff)
	lightest := uint8(0x00)
	hist := make([]uint32, 0xff+1)

	highest := uint8(0)
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

	}

	midpoint := darkest + (lightest-darkest)/2
	guess := (midpoint + highest) / 2
	/*
		fmt.Fprintf(os.Stderr, "mp: %d highest density %d\n", midpoint, highest)
		fmt.Fprintf(os.Stderr, "%+v\n", hist)
		fmt.Fprintf(os.Stderr, "%+v\n", gray.Pix)
	*/
	bw := segment.Threshold(gray, guess)
	return bw
	//	return gray
}
