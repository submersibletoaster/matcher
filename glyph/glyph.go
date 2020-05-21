package glyph

import (
	"image"
	"image/color"

	"github.com/pbnjay/pixfont"
)

// ThresholdPalette is a black/white color palette
var ThresholdPalette = color.Palette{color.RGBA{0, 0, 0, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff}}

// Match - scored match of font glyphs against an image cel
type Match struct {
	Score  float32 // Score - nearest to zero being closest match
	Char   string  // Char - string character of the matching font glyph
	Rune   rune    // Rune - codepoint of matching font glyph
	Invert bool    // Invert - indicates the match is for the inversion of the glyph bg<>fg
}

// Results - Sortable slice of Match
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

type GlyphInfo struct {
	image.Paletted
	uvHash []uint
}
type Lookup map[string]image.PalettedImage

type RasterFont struct {
	Font *pixfont.PixFont
	Lookup
}

// Query - Return scored matches of font glyphs which resemble the input image
// src Image is expected to be paletted with black at index 0 and
// white at index one (like ThresholdPalette)
func (RasterFont) Query(src *image.Paletted) (out Results) {

	return
}

// count set pixels and return the foreground and background density
// and the normalized foreground density - presumes src image is
// using ThresholdPalette
func imageDensity(src *image.Paletted) (fg, bg uint, norm float64) {
	for _, v := range src.Pix {
		fg += uint(v)
		if v == 0 {
			bg++
		}
	}
	norm = float64(fg) / float64(len(src.Pix))
	return
}

// uvHash - given a ThresholdPalette image returns a hash
// of density by column and row
func uvHash(src *image.Paletted) []uint {
	b := src.Bounds()
	column := make([]uint, b.Dx())
	row := make([]uint, b.Dy())
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			v := src.ColorIndexAt(x, y)
			column[x] += uint(v)
			row[y] += uint(v)
		}
	}
	uv := make([]uint, len(column)+len(row))
	for i, v := range column {
		uv[i] = v
	}
	for i, v := range row {
		uv[i+len(row)] = v
	}
	return uv
}
