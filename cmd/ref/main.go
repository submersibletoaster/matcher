package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/submersibletoaster/matcher/glyph"
	"github.com/submersibletoaster/matcher/unscii"
)

func main() {
	//	log.SetLevel(log.DebugLevel)
	font := unscii.Font

	// ASCII only
	/*
		chars := make([]rune, 0)

		for i := 0x20; i <= 0x7E; i++ {
			chars = append(chars, rune(i))
		}
	*/

	cm := unscii.CharMap()
	chars := make([]rune, 0)
	for k, _ := range cm {
		if glyph.UsablePoint(k) {
			chars = append(chars, rune(k))
		}
	}
	fmt.Printf("%d usable unicode codepoints\n", len(chars))

	g := glyph.NewRasterFont(font, chars)
	log.Infof("%+#v", g)
	perfect := 0
	edge := 0
	//chars = []rune{''}
	for _, c := range chars {
		r := g.Query(g.ImageForRune(c))
		//r = r[len(r)-5:len(r)-1]
		r = r[0:3]
		if r[0].Rune == c {
			perfect++
			continue
		}
		edge++
		fmt.Printf("'%s'\t%x\t", string(c), c)
		for _, v := range r {
			fmt.Printf("%.5f,'%s'\t", v.Score, string(v.Rune))
		}
		fmt.Println()
	}

	log.Infof("Perfect 1st match %d , edge cases %d\n", perfect, edge)

}
