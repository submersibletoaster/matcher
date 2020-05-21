package main

import (
	"fmt"
	"github.com/submersibletoaster/matcher/glyph"
	"github.com/submersibletoaster/matcher/unscii"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	font := unscii.Font
	
	
	/*
	chars := make([]rune,0)
	for i := 0x20 ; i <= 0x7E ; i++ {
		chars=append(chars,rune(i))
	}
	*/
	cm := unscii.CharMap()
	chars := make([]rune,len(cm))
	n := 0
	for k,_ := range cm {
		chars[n] = rune(k)
		n++
	}
	g := glyph.NewRasterFont(font,chars)
	log.Printf("%#+v\n",g)
	perfect := 0
	edge := 0
	for _,c := range chars {
		r := g.Query(g.ImageForRune(c))
		//r = r[len(r)-5:len(r)-1]
		r = r[0:3]
		if r[0].Rune == c {
			perfect++
			continue
		}
		edge++
		fmt.Printf("%s\t", string(c) )
		for _,v := range r {
			fmt.Printf("%.5f,'%s'\t",v.Score,string(v.Rune))
		}
		fmt.Println()
	}

	log.Infof("Perfect 1st match %d , edge cases %d\n", perfect, edge )

}
