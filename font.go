package matcher

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"sort"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/submersibletoaster/pixfont"
	"github.com/steakknife/hamming"
	"github.com/submersibletoaster/matcher/unscii"
)

const (
	version = 1.0
)

type Lookup map[string]image.PalettedImage

/*
type DStore struct {
	density [][]string
	lut     *Lookup
}

func (*DStore) Query(q *image.PalettedImage) []DMatch {
	d, nd := fgDensity(q)
	best := q.Bounds().Dx * q.Bounds().Dy

	for {

	}
}

type DMatch struct {
	score float32
	id    rune
}
*/

var myFont *pixfont.PixFont

var density [][]string
var lookup Lookup

func init() {
	myFont = unscii.Font
	lookup = fontMap(myFont)
	density = densityMap(lookup)
}

func GetLookup() Lookup {
	return lookup
}

func GetFont() *pixfont.PixFont {
	return myFont
}

func GetDensity() [][]string {
	return density
}

// fontMap generate a lookup of rune -> image of glyph
// which contains all the images keyed by their rune
func fontMap(font *pixfont.PixFont) Lookup {

	pal := color.Palette([]color.Color{color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}})
	lookup := Lookup{}
	bar := pb.StartNew(len(unscii.CharMap()))
	for r := range unscii.CharMap() {
		if usablePoint(r) {
			_, width := font.MeasureRune(rune(r))
			//img := image.NewRGBA(image.Rect(0, 0, width, font.GetHeight()))
			img := image.NewPaletted(image.Rect(0, 0, width, font.GetHeight()), pal)
			draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
			unscii.Font.DrawRune(img, 0, 0, rune(r), color.White)
			lookup[string(rune(r))] = img
		}
		bar.Increment()

	}
	bar.Finish()
	fmt.Fprintf(os.Stderr, "Used %d unscii runes\n", len(lookup))
	return lookup
}

func usablePoint(r int32) bool {
	// Borrowed from unscii/bm2uns-prebuild.pl selection of glyphs
	// (0x20)ASCII space
	// OR (0x2400 - 0x2bff ; representations of control chars, many drawing chars and line segments, arrows
	// 		bullets, circles, iching? lots of emoji
	// OR (0xe081 0xebff ) braille? and block shades, partial blocks, digital style  clock numerals
	// BUT NOT
	// 0x25fd,0x25fe filled/empty checkbox?
	// 0x2615 coffee emoji?
	// 0x26aa,0x26ab filled circles white,grey
	// 0x26f5 sailboat
	// 0x2b55 round red circle ?
	if (r == 0x20 || (r >= 0x2400 && r <= 0x2bff) || (r >= 0xe081 && r <= 0xebff)) && (r != 0x25fd && r != 0x25fe && r != 0x2615 && r != 0x26aa && r != 0x26ab && r != 0x26f5 && r != 0x2b55) {
		return true
	}
	return false

}

func densityMap(lut Lookup) [][]string {
	result := make([][]string, 256)
	count := 0
	for r, img := range lut {
		density := countPixels(img)
		if len(result[density]) == 0 {
			result[density] = make([]string, 0)
		}
		result[density] = append(result[density], r)
		count++
	}
	for d, runes := range result {
		if len(runes) == 0 {
			continue
		}
		fmt.Fprintf(os.Stderr, "%d\t%v\n", d, runes)
	}

	return result
}

func countPixels(img image.PalettedImage) (result uint32) {

	b := img.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := img.ColorIndexAt(x, y)
			result += uint32(c)
		}
	}
	return result
}

type Match struct {
	Score float32
	Char  string
}

type Results []Match

func (r Results) Swap(i, j int) {
	r[j], r[i] = r[i], r[j]
}
func (r Results) Less(i, j int) bool {
	return r[i].Score < r[j].Score
}
func (r Results) Len() int {
	return len(r)
}

func Closest(cel image.PalettedImage, chars []string) string {
	var matches Results
	max := float32(cel.Bounds().Dx() * cel.Bounds().Dy())
	for _, c := range chars {
		f := lookup[c]
		dist := hamming.Uint8s(cel.(*image.Paletted).Pix, f.(*image.Paletted).Pix)
		matches = append(matches, Match{float32(dist) / max, c})
	}
	sort.Sort(matches)
	fmt.Fprintf(os.Stderr, "%+v\n", matches)
	return matches[0].Char
}
