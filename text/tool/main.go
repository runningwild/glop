package main

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/runningwild/glop/text"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

var fontfile = flag.String("file", "", "Font file.")
var output = flag.String("output", "output", "Output filename prefix.")
var runes = flag.String("runes", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*(),.<>/?'\";:[]{}\\`~|", "Runes to render.")
var dpi = flag.Float64("dpi", 1000, "Dpi.")
var threads = flag.Int("threads", 4, "Number of threads to run simultaneously.")

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

	rIn := make(chan rune)
	riOut := make(chan runeImage)
	var wg sync.WaitGroup

	// Send all of the runes along rIn then close it.
	go func() {
		for _, r := range *runes {
			fmt.Printf("Processing %c\n", r)
			rIn <- r
		}
		close(rIn)
	}()

	// Bring up *threads goroutines.  Each one reads from rIn, processes the rune and sends a
	// runeImage along riOut.
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rIn {
				res, err := Render(font, r, *dpi, 0.05)
				if err != nil {
					fmt.Printf("Unable to render glyph: %v", err)
					os.Exit(1)
				}
				riOut <- runeImage{r: r, img: res}
			}
		}()
	}

	// Block until all of the above goroutines are done then close riOut.
	go func() {
		wg.Wait()
		close(riOut)
	}()

	// Take the results from riOut, put them into a slice and sort them.
	var ris runeImageSlice
	for ri := range riOut {
		ris = append(ris, ri)
	}
	sort.Sort(ris)

	width := 2048
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
	atlas.Set(0, 0, color.Gray{255})
	atlas.Set(0, 1, color.Gray{255})
	atlas.Set(1, 1, color.Gray{255})
	atlas.Set(2, 2, color.Gray{255})
	f, err := os.Create(fmt.Sprintf("%s.png", *output))
	if err != nil {
		fmt.Printf("Unable to make output file: %v", err)
		os.Exit(1)
	}
	err = png.Encode(f, atlas)
	if err != nil {
		fmt.Printf("Unable to encode png: %v", err)
		os.Exit(1)
	}
	f.Close()

	dict.Pix = atlas.Pix
	dict.Dx = int32(atlas.Bounds().Dx())
	dict.Dy = int32(atlas.Bounds().Dy())
	for _, r := range *runes {
		index := font.Index(r)
		ri := dict.Runes[r]
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
		dict.Runes[r] = ri
	}
	dict.Kerning = make(map[text.RunePair]int)
	for _, r0 := range *runes {
		for _, r1 := range *runes {
			kern := font.Kerning(font.FUnitsPerEm(), font.Index(r0), font.Index(r1))
			if kern == 0 {
				continue
			}
			dict.Kerning[text.RunePair{r0, r1}] = int(kern)
			fmt.Printf("Kern (%c, %c): %v\n", r0, r1, kern)
		}
	}
	dict.GlyphMax.Min.X = int(font.Bounds(font.FUnitsPerEm()).XMin)
	dict.GlyphMax.Min.Y = int(font.Bounds(font.FUnitsPerEm()).YMin)
	dict.GlyphMax.Max.X = int(font.Bounds(font.FUnitsPerEm()).XMax)
	dict.GlyphMax.Max.Y = int(font.Bounds(font.FUnitsPerEm()).YMax)

	{
		f, err := os.Create(fmt.Sprintf("%s.gob", *output))
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		enc := gob.NewEncoder(f)
		err = enc.Encode(dict)
		if err != nil {
			fmt.Printf("Failed to encode data to output file: %v\n", err)
			os.Exit(1)
		}
		f.Close()
	}
}
