package matcher

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"

	//"image/draw"

	"github.com/Nykakin/quantize"
	ansi "github.com/gookit/color"
	"github.com/joshdk/preview"
)

// Pick the overall image palette
func PickPalette(img image.Image, num int) color.Palette {
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

/*
func FindBestStructure(cell image.Image, store *duplo.Store) (*duplo.Match, color.Color, color.Color) {
	// github.com/disintegration/imaging
	//inverted := draw.Invert(cell)
	monoPal := PickPalette(cell, 2)
	//monoImg := image.NewPaletted(cell.Bounds(),monoPal)
	//draw.Draw(monoImg,cell.Bounds(),cell,image.ZP,draw.Src)
	monoImg := DitherToPalette(cell, monoPal, 2)

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
*/

func FindBestMatch(cell image.Image) (string, color.Color, color.Color) {
	dm := density
	monoPal := PickPalette(cell, 2)
	monoImg := QuantizeToPalette(cell, monoPal, 2)
	segGray := DynamicThreshold(cell)
	seg := image.NewPaletted(segGray.Bounds(), ThresholdPalette)
	draw.Draw(seg, seg.Bounds(), segGray, seg.Bounds().Min, draw.Src)
	monoBW := monoImg.(*image.Paletted)
	// *actual* representative colours
	bgCol := monoPal[0]
	fgCol := monoPal[1]

	// clobber with black/white to match the glyph database
	monoPal[0] = color.RGBA{0, 0, 0, 255}
	monoPal[1] = color.RGBA{255, 255, 255, 255}
	monoBW.Palette = monoPal
	cellFgDensity, limit, _ := fgDensity(seg)

	out := ""
	bg := bgCol
	fg := fgCol
	goal := cellFgDensity
	fmt.Fprintf(os.Stderr, "goal density: %v\n", goal)
	for {
		if len(dm[goal]) > 0 {
			out = Closest(seg, dm[goal])
			bg = bgCol
			fg = fgCol
			break
		} else if len(dm[limit-goal]) > 0 {
			out = Closest(seg, dm[limit-goal])
			bg = fgCol
			fg = bgCol
			break
		}
		goal--
		if goal < 0 {
			out = " "
			bg = bgCol
			fg = fgCol
			break
		}

	}

	//fmt.Fprintf(os.Stderr, "'%s'\n", out)
	return out, bg, fg
	//return rune(0x2318), bgCol, fgCol
}

func fgDensity(src *image.Paletted) (count uint, limit uint, norm float64) {
	limit = uint(src.Bounds().Dx() * src.Bounds().Dy())
	count = 0
	//fmt.Fprintf(os.Stderr, "%#v\n", src.Pix)
	for _, v := range src.Pix {
		//fmt.Fprintf(os.Stderr, "\tpixel: %v\n", v)
		count += uint(v)
	}
	return count, limit, float64(count) / float64(limit)
}

type RenderCB func(string, color.Color, color.Color)

func WriteANSI(w io.Writer, cells chan Cell, cb RenderCB) {
	for cell := range cells {
		char, bg, fg := FindBestMatch(cell.Image)
		cSeq := ansi.NewRGBStyle(toANSI(fg), toANSI(bg))

		if cell.Bounds.Min.X == 0 {
			//fmt.Fprint(w,"\n")
			fmt.Print("\033[0m\n")
		}
		cb(char, fg, bg)
		cSeq.Print(char)
		_ = cSeq
	}
}

func toANSI(in color.Color) (out ansi.RGBColor) {
	r, g, b, _ := in.RGBA()
	out = ansi.RGBColor{uint8(r), uint8(g), uint8(b), 0}
	return
}

/*
func PreviewImage(output draw.Image, cells chan Cell, cb RenderCB) {
	draw.Draw(output, output.Bounds(), image.Black, image.ZP, draw.Src)
	for cell := range cells {
		char, bg, fg := FindBestMatch(cell.Image)
		draw.Draw(output, cell.Bounds, image.NewUniform(bg), image.ZP, draw.Src)
		myFont.DrawRune(output, cell.Bounds.Min.X, cell.Bounds.Min.Y, rune(m.ID.(int32)), fg)
		cb(char, bg, fg)
	}
	preview.Image(output)
}
*/
