package matcher

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
)

func quantizeCell(cell image.Image, p color.Palette) {

}

type Cell struct {
	Image  image.Image
	Pos    image.Point
	Bounds image.Rectangle
}

func SliceImage(img image.Image, cell image.Rectangle, p color.Palette) (chan Cell, int) {
	size := img.Bounds()

	cellX := size.Dx() / cell.Dx()
	cellY := size.Dy() / cell.Dy()

	glyphX := cell.Dx()
	glyphY := cell.Dy()

	fmt.Fprintf(os.Stderr, "Image: %v\n", size)
	fmt.Fprintf(os.Stderr, "Cell: %v\n", cell)
	fmt.Fprintf(os.Stderr, "CellsXY: %d,%d\n", cellX, cellY)
	fmt.Fprintf(os.Stderr, "Width: %v\n", cell.Dx())

	cutFrom := img.(*image.RGBA)
	//cutFrom := image.NewPaletted(size,p)
	draw.Draw(cutFrom, cutFrom.Rect, img, size.Min, draw.Over)

	results := make(chan Cell, 8)
	go func() {
		yStep := glyphY
		xStep := glyphX
		for y := 0; y < size.Dy(); y += yStep {
			for x := 0; x < size.Dx(); x += xStep {
				origin := image.Rect(x, y, x+xStep, y+yStep)
				cel := cutFrom.SubImage(origin)
				outPos := image.Point{x / cell.Dx(), y / cell.Dy()}
				results <- Cell{Image: cel, Bounds: origin, Pos: outPos}
			}

		}
		close(results)
	}()
	return results, cellX * cellY
}
