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

	"github.com/submersibletoaster/matcher"
)

func main() {
	font := matcher.GetFont()
	lookup := matcher.GetLookup()
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

	func () {
	for y := 0; y < r.Max.Y; y += sY {
		for x := 0; x < r.Max.X; x += sX {
			font.DrawString(i, x, y, allChars[n], white)
			n++
			if n >= chars {
				return
			}
		}
	}
	}()

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

	wash := image.NewUniform(color.RGBA{0x88, 0x20, 0x40, 0xff})

	output := image.NewRGBA(srcImg.Bounds())
	draw.Draw(output, output.Bounds(), wash, image.ZP, draw.Src)

	thrOut := image.NewRGBA(srcImg.Bounds())
	draw.Draw(thrOut, output.Bounds(), wash, image.ZP, draw.Src)

	thrOutCol := image.NewRGBA(srcImg.Bounds())
	draw.Draw(thrOutCol, output.Bounds(), wash, image.ZP, draw.Src)

	pal := matcher.PickPalette(srcImg, 64)
	//hardpal := matcher.ThresholdPalette

	// TODO - be able to downsample rather than register the font glyph size directly as cell-size
	cells, _ := matcher.SliceImage(srcImg, image.Rect(0, 0, 8, 16), pal)
	for cell := range cells {
		/*		quant := matcher.QuantizeToPalette(cell.Image,pal,2)
				fmt.Fprintf(os.Stderr,"%+v\n", quant.(*image.Paletted).Palette )
				fmt.Fprintf(os.Stderr,"%+v\n", quant.(*image.Paletted).Pix )
				fmt.Fprintf(os.Stderr, "%#v\t\n", quant.Bounds())
				draw.Draw(output, cell.Bounds, quant, image.ZP  , draw.Src)
		*/
		seg := matcher.DynamicThreshold(cell.Image)
		fmt.Fprintf(os.Stderr, "Seg: %#v ; Cell: %v ; Pos: %v\n", seg.Bounds(), cell.Bounds, cell.Pos)
		fmt.Fprintf(os.Stderr, "%#v\n", seg)
		draw.Draw(thrOut, cell.Bounds, seg, seg.Bounds().Min, draw.Src)

		//celPal := matcher.PickPalette(cell.Image, 2)
		//seg.(*image.Paletted).Palette = celPal

		c, bg, fg := matcher.FindBestMatch(cell.Image)
		draw.Draw(output, cell.Bounds, image.NewUniform(bg), image.ZP, draw.Src)
		font.DrawString(output, cell.Bounds.Min.X, cell.Bounds.Min.Y, c, fg)

		//draw.Draw(thrOutCol, cell.Bounds, image.NewUniform(bg), image.ZP, draw.Src)
		mask := image.NewRGBA(image.Rect(0, 0, seg.Bounds().Dx(), seg.Bounds().Dy()))
		draw.Draw(mask, mask.Bounds(), seg, seg.Bounds().Min, draw.Src)
		maskPix := mask.Pix
		// the key is a mask has it's ALPHA channel used.
		for i := 0; i < len(maskPix); i += 4 {
			maskPix[i+3] = maskPix[i]
		}

		draw.Draw(thrOutCol, cell.Bounds, image.NewUniform(bg), cell.Bounds.Min, draw.Src)
		draw.DrawMask(thrOutCol, cell.Bounds, image.NewUniform(fg), cell.Bounds.Min,
			mask, mask.Bounds().Min, draw.Over)

	}

	ow, _ := os.Create("preview.png")
	png.Encode(ow, output)
	ow.Close()

	ot, _ := os.Create("threshold.png")
	png.Encode(ot, thrOut)
	ot.Close()

	otc, _ := os.Create("color-threshold.png")
	png.Encode(otc, thrOutCol)
	otc.Close()

}
