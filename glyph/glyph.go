package glyph

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/pbnjay/pixfont"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.Debug("")
}

// ThresholdPalette is a black/white color palette
var ThresholdPalette = color.Palette{color.RGBA{0, 0, 0, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff}}

// Match - scored match of font glyphs against an image cel
type Match struct {
	Score  float64 // Score - nearest to zero being closest match
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
	Image  *image.Paletted
	uvHash []uint
	r      rune
}
type Lookup map[string]image.PalettedImage

type RasterFont struct {
	Font      *pixfont.PixFont
	Width     int
	Height    int
	lutString map[string]*GlyphInfo
	lutRune   map[rune]*GlyphInfo
}

func NewRasterFont(f *pixfont.PixFont, chars []rune) (n RasterFont) {
	n.Font = f
	n.Width = f.MeasureString(" ")
	n.Height = f.GetHeight()
	n.makeInfo(chars)

	return n
}

func (s *RasterFont) makeInfo(chars []rune) {
	s.lutRune = make(map[rune]*GlyphInfo, len(chars))
	for _, r := range chars {
		img := image.NewPaletted(image.Rect(0, 0, s.Width, s.Height), ThresholdPalette)
		s.Font.DrawRune(img, 0, 0, r, image.White)
		hash := MakeUVHash(img)
		log.Debugf("%s\t%v\n", string(r), hash)
		g := GlyphInfo{img, hash, r}
		s.lutRune[r] = &g
	}
}

func (s RasterFont) ImageForRune(c rune) *image.Paletted {
	for img, ok := s.lutRune[c]; ok; {
		return img.Image
	}
	return image.NewPaletted(image.Rect(0, 0, s.Width, s.Height), ThresholdPalette)
}

// Query - Return scored matches of font glyphs which resemble the input image
// src Image is expected to be paletted with black at index 0 and
// white at index one (like ThresholdPalette)
func (s RasterFont) Query(src *image.Paletted) (out Results) {
	srcHash := MakeUVHash(src)
	for r, info := range s.lutRune {
		score := srcHash.CosineSimilarity(info.uvHash)
		//log.Debugf("%s\t%f\n", string(r), score)
		m := Match{Char: string(r), Rune: r, Score: score, Invert: false}
		out = append(out, m)
	}
	sort.Sort(sort.Reverse(out))
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

type uvHash []uint

func MakeUVHash(src *image.Paletted) uvHash {
	b := src.Bounds()
	column := make([]uint, b.Dx())
	row := make([]uint, b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := src.ColorIndexAt(x, y)
			column[x] += uint(v)
			row[y] += uint(v)
		}
	}
	uv := make(uvHash, len(column)+len(row))
	for i, v := range column {
		uv[i] = v
	}
	for i, v := range row {
		uv[i+len(column)-1] = v
		//		log.Debugf("\t%v\n", uv)
	}
	return uv
}

// uvHash could be compared by vector distance or
// hamming distance
func (a uvHash) CosineSimilarity(in uvHash) float64 {
	b := make(uvHash, len(a))
	for i, v := range in {
		b[i] = v
	}

	var numerator uint
	var aSq, bSq uint
	for i := 0; i < len(a); i++ {
		numerator += (a[i] * b[i])
		aSq += a[i] * a[i]
		bSq += b[i] * b[i]
	}
	log.Debugf("%d / (√%d * √%d)\n", numerator, aSq, bSq)
	similar := float64(numerator) / (math.Sqrt(float64(aSq)) * math.Sqrt(float64(bSq)))
	return similar
}
