package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	_ "golang.org/x/image/webp"

	ansi "github.com/gookit/color"

	"github.com/submersibletoaster/matcher/examine"
	"github.com/submersibletoaster/matcher/glyph"
	"github.com/submersibletoaster/pixfont"

	log "github.com/sirupsen/logrus"
	"github.com/submersibletoaster/matcher/unscii"
)

var font *pixfont.PixFont
var charmap map[rune]uint32
var rasterFont glyph.RasterFont

var workers = flag.Int("w", 1, "Number of worker routines")
var debugImages = flag.Bool("debug", false, "Output debug images for glyphs,cel thresholds and colored thresholds.")
var verbose = flag.Bool("v", false, "Verbose logging")
var width = flag.Int("x", 8, "X-Pixels per cell" )
var height = flag.Int("y",16, "Y-Pixels per cell")

func init() {
	flag.Parse()
	if *verbose {
		log.Info("Setting verbose logging")
		log.SetLevel(log.DebugLevel)
	}

	font = unscii.Font
	charmap = unscii.CharMap()
	chars := make([]rune, len(charmap))
	n := 0
	for k := range charmap {
		chars[n] = k
		n++
	}
	rasterFont = glyph.NewRasterFont(font, chars)
}

func main() {
	//makeCharSheet()

	srcFile := flag.Arg(0)
	srcIo, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	srcImg, _, _ := image.Decode(srcIo)
	if srcImg.ColorModel() != color.RGBAModel {
		old := srcImg
		replace := image.NewRGBA(old.Bounds())
		draw.Draw(replace, old.Bounds(), old, image.ZP, draw.Src)
		srcImg = replace
	}

	//cells := examine.ImageToCels(srcImg, 8, 16)
	cells := examine.ImageToCels(srcImg, *width, *height, 8, 16)


	renderers := make([]chan<- RenderOut, 0)
	toTerm := make(chan RenderOut, 1)
	go WriteANSI(os.Stderr, toTerm)
	renderers = append(renderers, toTerm)

	if *debugImages {
		toDebug := make(chan RenderOut, 1)
		go RenderDebugFunc(srcImg)(toDebug)
		renderers = append(renderers, toDebug)
	}

	Workers(uint(*workers), cells, renderers)

}

func makeCharSheet() {
	//font := matcher.GetFont()
	//lookup := matcher.GetLookup()
	lookup := rasterFont.GetLookup()
	chars := len(charmap)
	log.Debugf("%d chars\n", chars)
	cols := int(math.Ceil(math.Sqrt(float64(chars))))
	sX := 8 // magic number - should inspect from font
	sY := font.GetHeight()
	r := image.Rect(0, 0, sX*cols, sY*cols)
	i := image.NewRGBA(r)

	allChars := make([]string, chars)
	n := 0
	for g, _ := range lookup {
		allChars[n] = string(g)
		n++
	}

	n = 0
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	draw.Draw(i, i.Bounds(), image.NewUniform(black), image.ZP, draw.Src)

	func() {
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

}

func Workers(n uint, cels <-chan *examine.Cel, outputs []chan<- RenderOut) {
	mid := make(chan RenderOut, n)
	wait := sync.WaitGroup{}
	go func() {
		wait.Wait()
		close(mid)
	}()
	for i := uint(0); i < n; i++ {
		wait.Add(1)
		go func() {
			for cel := range cels {
				mid <- RenderOne(cel)
			}
			wait.Done()
		}()
	}

	// Sort the output from mid to ensure it goes out in CelPos order
	nextOut := uint(0)
	buffer := make(RenderBuff, 0)
	for o := range mid {
		buffer = append(buffer, o)
		sort.Sort(buffer)

		for len(buffer) != 0 && (buffer[0].Nth == nextOut) {
			for _, out := range outputs {
				out <- buffer[0]
			}
			nextOut++
			buffer = buffer[1:]
		}
	}
	for _, out := range outputs {
		close(out)
	}
	// FIXME kludge to let the Debug images save after its channel closes
	// should be a waitgroup of renderers maybe?
	time.Sleep(time.Second)

}

func RenderOne(cel *examine.Cel) RenderOut {
	seg, bg, fg := cel.DynamicThreshold()

	OP := rasterFont.DiffQuery

	results := OP(seg)
	//seg.Palette[0], seg.Palette[1] = seg.Palette[1], seg.Palette[0]
	inv := image.NewPaletted(seg.Bounds(), seg.Palette)
	inv.Palette = color.Palette{seg.Palette[1], seg.Palette[0]}
	draw.Draw(inv, seg.Bounds(), seg, seg.Bounds().Min, draw.Src)
	inv.Palette = seg.Palette

	resultsInverted := OP(inv)
	var c string
	var invert bool
	log.Debugf("CEL:%v", seg)
	if results[0].Score < resultsInverted[0].Score {
		log.Debugf("Regular match\t'%s'\t'%s'", results[0].Char, resultsInverted[0].Char)
		c = results[0].Char
		invert = false
	} else {
		c = resultsInverted[0].Char
		log.Debugf("Inverted char match\t'%s'\t'%s'", results[0].Char, resultsInverted[0].Char)
		fg, bg = bg, fg
		invert = true
	}

	log.Debugf("R\t%+v", results[0:3])
	log.Debugf("I\t%+v", resultsInverted[0:3])
	/*
		//seg.Palette[0], seg.Palette[1] = seg.Palette[1], seg.Palette[0]
		results := rasterFont.Query(seg)
		c := results[0].Char
	*/

	return RenderOut{c, fg, bg, cel.CharPos, cel.Nth, seg, invert}
}

func RenderDebugFunc(srcImg image.Image) func(<-chan RenderOut) {

	wash := image.NewUniform(color.RGBA{0x88, 0x20, 0x40, 0xff})
	output := image.NewRGBA(srcImg.Bounds())
	draw.Draw(output, output.Bounds(), wash, image.ZP, draw.Src)

	thrOut := image.NewRGBA(srcImg.Bounds())
	draw.Draw(thrOut, output.Bounds(), wash, image.ZP, draw.Src)

	thrOutCol := image.NewRGBA(srcImg.Bounds())
	draw.Draw(thrOutCol, output.Bounds(), wash, image.ZP, draw.Src)

	return func(r <-chan RenderOut) {
		log.Debugf("Starting debug out with chan %v", r)
		for ro := range r {
			seg, bg, fg,pos := ro.Segmented, ro.Bg, ro.Fg, ro.CelPos
			offset := image.Rect( pos.X * 8, pos.Y * 16 , (pos.X+1) * 8, (pos.Y+1) *16 )

			log.Debugf("Segment Bounds: %v", seg.Bounds())
			c := ro.Char
			draw.Draw(thrOut, offset, seg, seg.Bounds().Min, draw.Src)
			seg.Palette[0], seg.Palette[1] = seg.Palette[1], seg.Palette[0]
			draw.Draw(output, offset, image.NewUniform(bg), image.ZP, draw.Src)
			font.DrawString(output, offset.Bounds().Min.X, offset.Bounds().Min.Y, c, fg)

			mask := image.NewRGBA(image.Rect(0, 0, seg.Bounds().Dx(), seg.Bounds().Dy()))
			draw.Draw(mask, mask.Bounds(), seg, seg.Bounds().Min, draw.Src)
			maskPix := mask.Pix
			// the key is a mask has it's ALPHA channel used.
			if ro.Invert == true {
				for i := 0; i < len(maskPix); i += 4 {
					maskPix[i+3] = maskPix[i]
				}
			} else {
				for i := 0 ; i < len(maskPix); i += 4 {
					maskPix[i+3] = 255 - maskPix[i]
				}
			}
			//if ro.Invert == true {
			//	fg,bg = bg,fg
			//}
			draw.Draw(thrOutCol, seg.Bounds(), image.NewUniform(bg), seg.Bounds().Min, draw.Src)
			draw.DrawMask(thrOutCol, offset.Bounds(), image.NewUniform(fg), seg.Bounds().Min,
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
}

type RenderOut struct {
	Char      string
	Fg        color.Color
	Bg        color.Color
	CelPos    image.Point
	Nth       uint
	Segmented *image.Paletted
	Invert    bool
}

// RenderBuff - Sortable collection of RenderOut
// ANSI needs to emit results in cell order as a stream
//
type RenderBuff []RenderOut

func (r RenderBuff) Len() int {
	return len(r)
}
func (r RenderBuff) Less(i, j int) bool {
	return r[i].Nth < r[j].Nth
}
func (r RenderBuff) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func WriteANSI(w io.Writer, chars <-chan RenderOut) {
	for cel := range chars {
		cSeq := ansi.NewRGBStyle(toANSI(cel.Fg), toANSI(cel.Bg))

		if cel.CelPos.X == 0 {
			//fmt.Fprint(w,"\n")
			fmt.Print("\033[0m\n")
		}
		cSeq.Print(cel.Char)
		_ = cSeq
	}
	fmt.Fprintln(w)
}

func toANSI(in color.Color) (out ansi.RGBColor) {
	r, g, b, _ := in.RGBA()
	out = ansi.RGBColor{uint8(r), uint8(g), uint8(b), 0}
	return
}
