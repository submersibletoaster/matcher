package main

import (
	"flag"
	"image"
        "image/color"
	"image/png"
	"image/draw"
	"os"
	"sort"
	"fmt"
	pb "github.com/cheggaaa/pb/v3"
	"github.com/pbnjay/pixfont"
	"github.com/submersibletoaster/matcher/unscii"

	"github.com/rivo/duplo"


)

var cellX = flag.Int("w",8,"Cell width")
var cellY = flag.Int("h",8,"Cell height")

func main() {
	flag.Parse()

	// 
	srcFile := flag.Arg(0)
	srcIo,err := os.Open(srcFile)
	srcImg,_,err := image.Decode(srcIo)
	if err != nil {
		panic(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 150, 30))
	draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
	unscii.Font.DrawString(img, 10, 10, "Hello, World!", color.White)
	//pixfont.DrawString(img, 10, 10, "Hello, World!", color.White)
        f, _ := os.OpenFile("hello.png", os.O_CREATE|os.O_RDWR, 0644)
        png.Encode(f, img)
	fmt.Println("matcher…")

	_,store := fontMap(unscii.Font)
	srcHash,_ := duplo.CreateHash(srcImg)
	matches := store.Query(srcHash)
	sort.Sort(matches)
	fmt.Printf("Top match:\t%+v\n", matches[0])

}

type Lookup map[rune]*image.RGBA

func fontMap(font *pixfont.PixFont) (Lookup,*duplo.Store) {
	lookup := Lookup{}
        store := duplo.New()
	bar := pb.StartNew(len(unscii.CharMap))
	for r,_ := range unscii.CharMap {
		_,width := font.MeasureRune(rune(r))
		img := image.NewRGBA(image.Rect(0,0,width,font.GetHeight()))
		draw.Draw(img,img.Bounds(),image.Black,image.ZP,draw.Src)
		unscii.Font.DrawRune(img,0,0,rune(r),color.White)
		hash,_ := duplo.CreateHash(img)
		store.Add(rune(r),hash)
		lookup[rune(r)] = img
		bar.Increment()
	}
	bar.Finish()
	return lookup,store
}


