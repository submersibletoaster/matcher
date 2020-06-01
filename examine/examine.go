package examine

import (
	"image"
	"image/color"
	"image/draw"
	"sort"

	//"github.com/Nykakin/quantize"
	"github.com/ericpauley/go-quantize/quantize"

	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
)

// White - reference White colorful
var White colorful.Color

// Black - reference Balck colorful
var Black colorful.Color

func init() {
	White, _ = colorful.MakeColor(color.White)
	Black, _ = colorful.MakeColor(color.Black)
}

// ImageToCels - Take a source image and slice into subimages
// of cellX,cellY
func ImageToCels(src image.Image, cellX int, cellY int) <-chan *Cel {
	b := src.Bounds()
	//width := b.Dx()
	//height := b.Dy()
	out := make(chan *Cel, 1)
	copy := image.NewRGBA(b)
	draw.Draw(copy, b, src, b.Min, draw.Src)
	go func() {
		nth := uint(0)
		for y := b.Min.Y; y <= b.Max.Y-cellY; y += cellY {
			for x := b.Min.X; x <= b.Max.X-cellX; x += cellX {
				origin := image.Rect(x, y, x+cellX, y+cellY)
				cel := copy.SubImage(origin).(*image.RGBA)
				charPos := image.Point{x / cellX, y / cellY}
				//log.Debugf("ImageToCells: %v", charPos)
				out <- &Cel{Image: cel, Origin: origin, CharPos: charPos, Nth: nth}
				nth++
			}
		}
		log.Debug("ImageToCells: Closing channel")
		close(out)
	}()
	return out
}

type Cel struct {
	Image   *image.RGBA
	Origin  image.Rectangle
	CharPos image.Point
	Nth     uint
}

type LabColors []colorful.Color

// Getting the right fit for contrasting colors is a real struggle.
// just using a quantizer to crush to 2 seems best so far.
// - this tends to loose structure and contrast
// better to choose more colors and be selective about
// the contrast

type Contrast struct {
	Color    colorful.Color
	Distance float64
}

type Contrasting []Contrast

func (c Contrasting) Less(i, j int) bool {
	return c[i].Distance < c[j].Distance
}
func (c Contrasting) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Contrasting) Len() int {
	return len(c)
}

func (s Cel) ContrastingColors() []color.Color {
	cols := make([]color.Color, 0, 8)
	q := quantize.MedianCutQuantizer{}
	p := q.Quantize(cols, s.Image)

	if len(p) < 2 {
		return []color.Color{p[0], p[0]}
		panic("Palette is single color")
	}

	// Naive - most prevalent and least prevalent palette colors
	out := []color.Color{p[0], p[len(p)-1]}
	return out

	// Choose the two colors with the greatest distance from others
	cf := make(Contrasting, len(p))
	// for i, c := range p {
	// 	col, _ := colorful.MakeColor(c)
	// 	dist := 0.0
	// 	for _, d := range p {
	// 		dc, _ := colorful.MakeColor(d)
	// 		dist += col.DistanceLab(dc)
	// 	}
	// 	cf[i] = Contrast{col, dist}

	// }
	// sort.Sort(sort.Reverse(cf))
	for i, c := range p {
		col, _ := colorful.MakeColor(c)
		cf[i] = Contrast{col, 0.0}
	}

	primary, cf := cf[0], cf[1:]
	for _, c := range cf {
		c.Distance = c.Color.DistanceLab(primary.Color)
	}
	sort.Sort(sort.Reverse(cf))

	return []color.Color{primary.Color, cf[0].Color}
}

// ContrastingColors - slice of lightest and darkest seen colors
func (s Cel) NOTContrastingColors() []colorful.Color {
	b := s.Image.Bounds()

	dark := make(map[colorful.Color]uint)
	light := make(map[colorful.Color]uint)
	distinct := make(map[colorful.Color]uint)
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			lab, _ := colorful.MakeColor(s.Image.At(x, y))
			dBright := lab.DistanceLab(White)
			dDark := lab.DistanceLab(Black)
			// Histogram lighter and darker colors
			// equal distance from black white ignored
			distinct[lab]++
			if dDark < dBright {
				dark[lab]++
			} else if dDark > dBright {
				light[lab]++
			}
		}
	}
	log.Debugf("Distinct colors: %d", len(distinct))
	out := make([]colorful.Color, 2)

	lightMax := uint(0)
	for k, v := range light {
		if v > lightMax {
			out[0] = k
		}
	}
	darkMax := uint(0)
	for k, v := range dark {
		if v > darkMax {
			out[1] = k
		}
	}
	if out[0] == out[1] {
		log.Fatalf("Contrasting colors are same: %v\n%v\n\tLight %v\n\tDark %v\n", out, distinct, light, dark)
	}
	return out
}

// DynamicThreshold - reduce the cels image to a two color paletted image,
func (s Cel) DynamicThreshold() (*image.Paletted, color.Color, color.Color) {
	cols := s.ContrastingColors()
	origPal := color.Palette{cols[0], cols[1]}
	pal := make(color.Palette, 2)
	pal[0] = Black
	pal[1] = White
	out := image.NewPaletted(s.Origin, origPal)
	draw.FloydSteinberg.Draw(out, s.Origin, s.Image, s.Origin.Min)
	//draw.Draw(out, s.Origin, s.Image, s.Origin.Min, draw.Src)

	out.Palette = pal
	return out, cols[0], cols[1]
}

/*
func (s Cel) GetDominantColors() []*color.RGBA {

}

func (s Cel) Threshold() uint {

}
*/
