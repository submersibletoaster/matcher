package matcher

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"

	pb "github.com/cheggaaa/pb/v3"
	"github.com/pbnjay/pixfont"
	"github.com/rivo/duplo"
	"github.com/submersibletoaster/matcher/unscii"
)

var myFont *pixfont.PixFont

// duplo image similarity store
var fontStore *duplo.Store

func init() {
	myFont = unscii.Font
	_, fontStore = fontMap(myFont)
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
		// Borrowed from unscii/bm2uns-prebuild.pl selection of glyphs
		if (r == 0x20 || (r >= 0x2400 && r <= 0x2bff) || (r >= 0xe081 && r <= 0xebff)) && (r != 0x25fd && r != 0x25fe && r != 0x2615 && r != 0x26aa && r != 0x26ab && r != 0x26f5 && r != 0x2b55) {

			_, width := font.MeasureRune(rune(r))
			img := image.NewRGBA(image.Rect(0, 0, width, font.GetHeight()))
			draw.Draw(img, img.Bounds(), image.Black, image.ZP, draw.Src)
			unscii.Font.DrawRune(img, 0, 0, rune(r), color.White)
			if !haveCached {
				hash, _ := duplo.CreateHash(img)
				store.Add(rune(r), hash)
			}
			lookup[rune(r)] = img
		}
		bar.Increment()

	}
	bar.Finish()
	fmt.Fprintf(os.Stderr, "Used %d unscii runes\n", len(lookup))
	return lookup, store
}
