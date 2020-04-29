package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

func quantizeCell(cell image.Image, p color.Palette) {
	
}

func sliceImage(img image.Image, cell image.Rectangle,p color.Palette) chan image.Image {
	size := img.Bounds()

	cellX := size.Dx() / cell.Dx()
	cellY := size.Dy() / cell.Dy()

	fmt.Printf("Image: %v\n", size)
	fmt.Printf("Cell: %v\n", cell)
	fmt.Printf("CellXY: %d,%d\n",cellX,cellY)
	fmt.Printf("Width: %v\n", cell.Dx())

	//cutFrom := img.(*image.RGBA)

	cutFrom := image.NewPaletted(size,p)
	draw.Draw(cutFrom, cutFrom.Rect, img, size.Min, draw.Over)


	results := make(chan image.Image)
	go func() {
		for x := 0; x < size.Dx(); x += cellX {
			for y := 0; y < size.Dy(); y += cellY {
				cel := cutFrom.SubImage(image.Rect(x, y, x+cell.Dx(), y+cell.Dy() ))
				results <- cel
			}

		}
		close(results)
	}()
	return results
}

