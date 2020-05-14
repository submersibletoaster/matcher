package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"io/ioutil"
	"os"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/pbnjay/pixfont"
	"github.com/rivo/duplo"
	"github.com/submersibletoaster/matcher/unscii"
)

var cellX = flag.Int("w", 8, "Cell width")
var cellY = flag.Int("h", 16, "Cell height")

var myFont *pixfont.PixFont
var fontStore *duplo.Store
func init(){
	myFont = unscii.Font
	_, fontStore = fontMap(myFont)
}

func main() {
	flag.Parse()

	srcFile := flag.Arg(0)
	srcIo, err := os.Open(srcFile)
	srcImg, _, err := image.Decode(srcIo)
	if err != nil {
		panic(err)
	}
	if srcImg.ColorModel() != color.RGBAModel {
		old := srcImg
		replace := image.NewRGBA(old.Bounds())
		draw.Draw(replace, old.Bounds(), old, image.ZP, draw.Src)
		srcImg = replace
	}

	//img := image.NewRGBA(image.Rect(0, 0, 150, 30))
	//draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
	//unscii.Font.DrawString(img, 10, 10, "Hello, World!", color.White)
	//pixfont.DrawString(img, 10, 10, "Hello, World!", color.White)
	//f, _ := os.OpenFile("hello.png", os.O_CREATE|os.O_RDWR, 0644)
	//png.Encode(f, img)

	fmt.Println("matcher…")
	fmt.Printf("duplo.ImageScale is %d\n", duplo.ImageScale)

	pal := pickPalette(srcImg, 64)

	// TODO - be able to downsample rather than register the font glyph size directly as cell-size
	cells, expectCells := sliceImage(srcImg, image.Rect(0, 0, *cellX, *cellY), pal)


	bar := pb.StartNew(expectCells)
	output := image.NewRGBA(srcImg.Bounds())
	diverse := make(map[string]uint)
	cb := func(c string,bg color.Color,fg color.Color) {
		bar.Increment()
		diverse[c]++
	}
	previewImage(output,cells,cb)

	/*
	for cell := range cells {
		m, bg, fg := findBestStructure(cell.Image, fontStore)
		char := string(rune(m.ID.(int32)))
		diverse[char]++
		draw.Draw(output, cell.Bounds, image.NewUniform(bg), image.ZP, draw.Src)
		myFont.DrawRune(output, cell.Bounds.Min.X, cell.Bounds.Min.Y, rune(m.ID.(int32)), fg)
		//draw.Draw(output,cell.Bounds,cell.Image,cell.Image.Bounds().Min,draw.Src)
		bar.Increment()
	}
	bar.Finish()
	preview.Image(output)
	fmt.Printf("Diversity: %+v", diverse)
	*/
}

type Lookup map[rune]*image.RGBA

// fontMap generate a lookup of rune -> image of glyph and a duplo.Store
//  (possibly cached) which contains all the images keyed by their rune
func fontMap(font *pixfont.PixFont) (Lookup, *duplo.Store) {

	// We can cache these based on font not changing
	haveCached := false
	cacheFile, err := os.Open(".duplo.cache")
	store := duplo.New()
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No duplo cache for matching: %s\n", err)
		defer func() {
			out, err := store.GobEncode()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write duplo cache: %s", err)
			} else {
				ioutil.WriteFile(".duplo.cache", out, 0664)
			}
		}()
	} else {
		in, err := ioutil.ReadAll(cacheFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open duplo cache: %s", err)
		} else {
			store.GobDecode(in)
			haveCached = true
		}
	}

	lookup := Lookup{}
	bar := pb.StartNew(len(unscii.CharMap))
	for r := range unscii.CharMap {
		_, width := font.MeasureRune(rune(r))
		img := image.NewRGBA(image.Rect(0, 0, width, font.GetHeight()))
		draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
		unscii.Font.DrawRune(img, 0, 0, rune(r), color.White)
		if !haveCached {
			hash, _ := duplo.CreateHash(img)
			store.Add(rune(r), hash)
		}
		lookup[rune(r)] = img
		bar.Increment()
	}
	bar.Finish()
	return lookup, store
}
