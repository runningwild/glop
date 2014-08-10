// Binary tool generates dictionary files that can be used by the text package to render text using
// the distance field font rendering method.  A .ttf font file must be specified as an input, and an
// output.png and output.dict file will be generated from it.  The output.png file can be used to
// see how well packed the atlas is, if there is a significant amount of empty space you may want to
// retry with a different value for --dpi.  Regardless of what --dpi is used when generating the
// dictionary the font will always render at the same size because a height parameter is always used
// when rendering text.  The output.dict file is what should be given to text.LoadDictionary().
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
var runes = flag.String("runes", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*(),.<>/?'\";:[]{}\\`~| ", "Runes to render, 'ALL' to use all avialable runes (< 65535) in this font.")
var dpi = flag.Float64("dpi", 300, "Dpi.  300 is usually sufficient to look good at all reasonable sizes.")
var threads = flag.Int("threads", 4, "Number of threads to run simultaneously.")
var width = flag.Int("width", 1024, "Width of atlas.")

// runeImage is a simple pair of rune and the grayscale image of that rune.
type runeImage struct {
	r   rune
	img *image.Gray
}

// runeImageSlice is used to sort runeImages to make packing easier.
type runeImageSlice []runeImage

func (ris runeImageSlice) Len() int {
	return len(ris)
}
func (ris runeImageSlice) Less(i, j int) bool {
	if ris[i].img.Bounds().Dy() != ris[j].img.Bounds().Dy() {
		return ris[i].img.Bounds().Dy() > ris[j].img.Bounds().Dy()
	}
	return ris[i].img.Bounds().Dx() < ris[j].img.Bounds().Dx()
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

	all := *runes == "ALL"
	var usable []rune
	fmt.Printf("This font supports the following runes:\n")
	fmt.Printf("Rune, Val, Index\n")
	var gCheck truetype.GlyphBuf
	for r := rune(1); r < 65535; r++ {
		index := font.Index(r)
		if index == 0 {
			continue
		}
		err := gCheck.Load(font, 1, index, truetype.NoHinting)
		if err == nil {
			fmt.Printf("'%c': %d, %d\n", r, r, index)
		}
		if all {
			usable = append(usable, r)
		}
	}
	if all {
		*runes = string(usable)
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

	// Now that runes are sorted, pack them.  The packing is simple, just fill up a row until the
	// next rune wouldn't fit, then go to the next row.  The runes are sorted first by largest
	// height then by smallest width, which isn't necessarily optimal, but it is good in general.
	// The +2 on several values here is so that there is a small buffer between runes so that we
	// don't accidentally draw part of one rune while render another.
	width := *width
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
	height := 1
	for height < y {
		height *= 2
	}

	// Now that we know the position of all the runes in the atlas, actually draw them onto a single
	// image.
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

	// Save the output.png file so that we can see what the atlas looks like.
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

	// There are several more values that need to go into a dictionary, so we'll go through all of
	// the runes that we're using and look up those values now.
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

	// Also store a mapping from rune pair to the kerning for that pair, if this font has that info.
	dict.Kerning = make(map[text.RunePair]int)
	for _, r0 := range *runes {
		for _, r1 := range *runes {
			kern := font.Kerning(font.FUnitsPerEm(), font.Index(r0), font.Index(r1))
			if kern == 0 {
				continue
			}
			dict.Kerning[text.RunePair{r0, r1}] = int(kern)
		}
	}
	dict.GlyphMax.Min.X = int(font.Bounds(font.FUnitsPerEm()).XMin)
	dict.GlyphMax.Min.Y = int(font.Bounds(font.FUnitsPerEm()).YMin)
	dict.GlyphMax.Max.X = int(font.Bounds(font.FUnitsPerEm()).XMax)
	dict.GlyphMax.Max.Y = int(font.Bounds(font.FUnitsPerEm()).YMax)

	// Now store the output.dict file so we can use it with text.LoadDictionary()
	{
		f, err := os.Create(fmt.Sprintf("%s.dict", *output))
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
