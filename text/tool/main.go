package main

import (
	"bytes"
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"flag"
	"fmt"
	"github.com/runningwild/glop/text"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

var fontfile = flag.String("file", "", "Font file.")
var runes = flag.String("runes", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*(),.<>/?'\";:[]{}\\`~|", "Runes to render.")
var dpi = flag.Float64("dpi", 1000, "Dpi.")

type runeImage struct {
	r   rune
	img *image.Gray
}
type runeImageSlice []runeImage

func (ris runeImageSlice) Len() int {
	return len(ris)
}
func (ris runeImageSlice) Less(i, j int) bool {
	return ris[i].img.Bounds().Dy() > ris[j].img.Bounds().Dy()
}
func (ris runeImageSlice) Swap(i, j int) {
	ris[i], ris[j] = ris[j], ris[i]
}

func main() {
	flag.Parse()
	data, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		fmt.Printf("Unable to read file: %v", err)
		os.Exit(1)
	}
	font, err := freetype.ParseFont(data)
	if err != nil {
		fmt.Printf("Unable to parse font file: %v", err)
		os.Exit(1)
	}
	var ris runeImageSlice
	for _, r := range *runes {
		fmt.Printf("Processing %c\n", r)
		res, err := Render(font, r, *dpi, 0.05)
		if err != nil {
			fmt.Printf("Unable to render glyph: %v", err)
			os.Exit(1)
		}
		ris = append(ris, runeImage{r: r, img: res})
	}

	sort.Sort(ris)
	width := 1024
	x, y := 0, 0
	cdy := ris[0].img.Bounds().Dy() + 2
	for _, ri := range ris {
		dx := ri.img.Bounds().Dx() + 2
		if x+dx > width {
			x = 0
			y += cdy
			cdy = ri.img.Bounds().Dy() + 2
		}
		x += dx
	}
	y += cdy
	fmt.Printf("Finished with %d %d\n", x, y)
	height := 1
	for height < y {
		height *= 2
	}
	atlas := image.NewGray(image.Rect(0, 0, width, height))
	x, y = 0, 0
	cdy = ris[0].img.Bounds().Dy() + 2
	var dict text.Dictionary
	dict.Runes = make(map[rune]text.RuneInfo)
	for _, ri := range ris {
		dx := ri.img.Bounds().Dx() + 2
		if x+dx > width {
			x = 0
			y += cdy
			cdy = ri.img.Bounds().Dy() + 2
		}
		dict.Runes[ri.r] = text.RuneInfo{PixBounds: ri.img.Bounds().Add(image.Point{x + 1, y + 1})}
		draw.Draw(atlas, dict.Runes[ri.r].PixBounds, ri.img, image.Point{}, draw.Over)
		x += dx
	}
	f, err := os.Create("output.png")
	if err != nil {
		fmt.Printf("Unable to make output file: %v", err)
		os.Exit(1)
	}
	var buf bytes.Buffer
	err = png.Encode(io.MultiWriter(f, &buf), atlas)
	if err != nil {
		fmt.Printf("Unable to encode png: %v", err)
		os.Exit(1)
	}
	f.Close()

	dict.Pix = buf.Bytes()
	for _, r := range *runes {
		index := font.Index(r)
		var ri text.RuneInfo
		ri.AdvanceHeight = int(font.VMetric(font.FUnitsPerEm(), index).AdvanceHeight)
		ri.TopSideBearing = int(font.VMetric(font.FUnitsPerEm(), index).TopSideBearing)
		ri.AdvanceWidth = int(font.HMetric(font.FUnitsPerEm(), index).AdvanceWidth)
		ri.LeftSideBearing = int(font.HMetric(font.FUnitsPerEm(), index).LeftSideBearing)
		glyph := truetype.NewGlyphBuf()
		glyph.Load(font, font.FUnitsPerEm(), index, truetype.FullHinting)
		ri.GlyphBounds.Min.X = int(glyph.B.XMin)
		ri.GlyphBounds.Min.Y = int(glyph.B.YMin)
		ri.GlyphBounds.Max.X = int(glyph.B.XMax)
		ri.GlyphBounds.Max.Y = int(glyph.B.YMax)
	}
	dict.Kerning = make(map[text.RunePair]int)
	for _, r0 := range *runes {
		for _, r1 := range *runes {
			dict.Kerning[text.RunePair{r0, r1}] = int(font.Kerning(font.FUnitsPerEm(), font.Index(r0), font.Index(r1)))
		}
	}
}
