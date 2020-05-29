package examine

import (
	"image"
	"image/color"
	"image/draw"

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
		for y := b.Min.Y; y < b.Max.Y-cellY; y += cellY {
			for x := b.Min.X; x < b.Max.X-cellX; x += cellX {
				origin := image.Rect(x, y, x+cellX, y+cellY)
				cel := copy.SubImage(origin).(*image.RGBA)
				charPos := image.Point{x / cellX, y / cellY}
				log.Debugf("ImageToCells: %v", charPos)
				out <- &Cel{Image: cel, Origin: origin, CharPos: charPos}
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
}

type LabColors []colorful.Color

// ContrastingColors - slice of lightest and darkest seen colors
func (s Cel) ContrastingColors() []colorful.Color {
	b := s.Image.Bounds()

	dark := make(map[colorful.Color]uint)
	light := make(map[colorful.Color]uint)
	for y := b.Min.Y; y <= b.Max.Y; y++ {
		for x := b.Min.X; x <= b.Max.X; x++ {
			lab, _ := colorful.MakeColor(s.Image.At(x, y))
			dBright := lab.DistanceLab(White)
			dDark := lab.DistanceLab(Black)
			// Histogram lighter and darker colors
			// equal distance from black white ignored
			if dDark < dBright {
				dark[lab]++
			} else if dDark > dBright {
				light[lab]++
			}
		}
	}

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
	return out
}

func (s Cel) DynamicThreshold() (*image.Paletted, color.Color, color.Color) {
	cols := s.ContrastingColors()
	pal := make(color.Palette, 2)
	pal[0] = cols[0]
	pal[1] = cols[1]
	out := image.NewPaletted(s.Origin, pal)
	draw.Draw(out, s.Origin, s.Image, s.Origin.Min, draw.Src)
	return out, cols[0], cols[1]
}

/*
func (s Cel) GetDominantColors() []*color.RGBA {

}

func (s Cel) Threshold() uint {

}
*/
